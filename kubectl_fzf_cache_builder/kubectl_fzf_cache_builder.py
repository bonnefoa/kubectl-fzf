#!env python2
import argparse
from kubernetes import client, config, watch
from kubernetes.config import kube_config
import logging
import os.path
import multiprocessing
import signal
import subprocess
import time
from . import resource
import yaml
import base64
import json
import datetime
from dateutil import tz
from urllib3.exceptions import ProtocolError


log = logging.getLogger('dd.' + __name__)

watches = []
exiting = False


def is_expired(user):
    if 'auth-provider' not in user:
        return False
    provider = user['auth-provider']
    if 'config' not in provider:
        return False
    parts = provider['config']['id-token'].split('.')
    if len(parts) != 3:
        return False
    jwt_attributes = json.loads(base64.b64decode(parts[1] + "=="))
    expire = jwt_attributes.get('exp')
    if expire is None:
        return False
    if kube_config._is_expired(
            datetime.datetime.fromtimestamp(expire, tz=tz.tzutc())):
        return True
    return False


def get_kubernetes_config(cluster, refresh_command):
    res = None

    if refresh_command:
        filename = os.path.expanduser(kube_config.KUBE_CONFIG_DEFAULT_LOCATION)
        with open(filename) as f:
            loader = kube_config.KubeConfigLoader(
                config_dict=yaml.load(f),
                config_base_path=os.path.abspath(os.path.dirname(filename)))
            if is_expired(loader._user):
                subprocess.call([refresh_command, cluster])

    res = config.new_client_from_config(context=cluster)
    return res


class ResourceDumper(object):

    def __init__(self, dest_dir, resource_cls):
        self.resource_cls = resource_cls
        self.header = resource_cls.header()
        self.dest_file = os.path.join(dest_dir, resource_cls._dest_file())
        self.tmp_file = '{}_'.format(self.dest_file)
        self.f = open(self.dest_file, 'w')

    def write_resources_to_file(self, resources):
        self.f.write('{}\n'.format(self.header))
        self.f.writelines(['{}\n'.format(str(r)) for r in resources])
        self.f.flush()

    def write_resource_to_file(self, resource, resource_dict, truncate_file):
        if truncate_file:
            self.f.close()
            self.f = open('{}_'.format(self.dest_file), 'w')
            self.write_resources_to_file(resource_dict)
            os.rename(self.tmp_file, self.dest_file)
        else:
            if self.f.tell() == 0:
                self.f.write('{}\n'.format(self.header))
            self.f.write('{}\n'.format(str(resource)))
            self.f.flush()

    def close(self):
        self.f.close()
        self.f = None


class ResourceWatcher(object):

    def __init__(self, cluster, namespace, args):
        self.cluster = cluster
        self.namespace = namespace
        self.refresh_command = args.refresh_command
        self.node_poll_time = args.node_poll_time
        self.namespace_poll_time = args.namespace_poll_time
        self.kube_kwargs = {'_request_timeout': 600}
        self.dir = args.dir
        if args.selector:
            self.kube_kwargs['label_selector'] = args.selector
        if self.namespace != 'all':
            self.kube_kwargs['namespace'] = self.namespace
        self.construct_kubernetes_api()

    def construct_kubernetes_api(self):
        kubernetes_config = get_kubernetes_config(self.cluster,
                                                  self.refresh_command)
        self.v1 = client.CoreV1Api(api_client=kubernetes_config)
        self.apps_v1 = client.AppsV1Api(api_client=kubernetes_config)
        self.extensions_v1beta1 = client.ExtensionsV1beta1Api(api_client=kubernetes_config)

    def process_resource(self, resource, resource_dict, resource_dumper):
        do_truncate = False
        if resource.is_deleted and resource in resource_dict:
            log.debug('Removing resource {}'.format(resource))
            resource_dict.pop(resource)
            do_truncate = True
        elif not resource.is_deleted:
            if resource in resource_dict and \
                    str(resource) != str(resource_dict[resource]):
                log.debug('Updating resource {}'.format(resource))
                do_truncate = True
                resource_dict.pop(resource)
            resource_dict[resource] = resource
        resource_dumper.write_resource_to_file(resource, resource_dict, do_truncate)

    def _get_resource_kwargs(self, Resource):
        kwargs = self.kube_kwargs
        if not Resource._has_namespace():
            kwargs = dict(self.kube_kwargs)
            kwargs.pop('namespace', None)
        return kwargs

    def watch_resource(self, func, ResourceCls):
        resource_dumper = ResourceDumper(self.dir, ResourceCls)
        log.warn('Watching {} on namespace {}, writing results in {}'.format(
            ResourceCls.__name__, self.namespace, resource_dumper.dest_file))
        w = watch.Watch()
        watches.append(w)
        resource_dict = {}
        kwargs = self._get_resource_kwargs(ResourceCls)
        i = 0
        try:
            for resp in w.stream(func, **kwargs):
                resource = ResourceCls(resp['object'])
                self.process_resource(resource, resource_dict, resource_dumper)
                i = i + 1
                if i % 1000 == 0:
                    log.info('Processed {} {}'.format(i, ResourceCls.__name__))
            resource_dumper.close()
        except ProtocolError as e:
            log.warn('{} watcher exiting due to {}'.format(ResourceCls.__name__, e))

    def poll_resource(self, func, poll_time, ResourceCls):
        dest_file=os.path.join(self.dir, ResourceCls._dest_file())
        log.info('Poll {} on namespace {}, writing results in {}'.format(
            ResourceCls.__name__, self.namespace, dest_file))
        kwargs = self._get_resource_kwargs(ResourceCls)
        resource_dumper = ResourceDumper(self.dir, ResourceCls)
        while exiting is False:
            resp = func(**kwargs)
            resources = [ResourceCls(item) for item in resp.items]
            resource_dumper.write_resources_to_file(resources)
            time.sleep(poll_time)
        resource_dumper.close()
        log.info('{} poll exiting {}'.format(ResourceCls.__name__, exiting))

    def watch_pods(self):
        func = self.v1.list_namespaced_pod
        if self.namespace == 'all':
            func = self.v1.list_pod_for_all_namespaces
        self.watch_resource(func, resource.Pod)

    def watch_services(self):
        func = self.v1.list_namespaced_service
        if self.namespace == 'all':
            func = self.v1.list_service_for_all_namespaces
        self.watch_resource(func, resource.Service)

    def watch_nodes(self):
        self.poll_resource(self.v1.list_node, self.node_poll_time, resource.Node)

    def watch_replicaset(self):
        func = self.apps_v1.list_namespaced_replica_set
        if self.namespace == 'all':
            func = self.apps_v1.list_replica_set_for_all_namespaces
        self.watch_resource(func, resource.ReplicaSet)

    def watch_configmap(self):
        func = self.v1.list_namespaced_config_map
        if self.namespace == 'all':
            func = self.v1.list_config_map_for_all_namespaces
        self.watch_resource(func, resource.ConfigMap)

    def watch_endpoint(self):
        func = self.v1.list_namespaced_endpoints
        if self.namespace == 'all':
            func = self.v1.list_endpoints_for_all_namespaces
        self.watch_resource(func, resource.Endpoint)

    def watch_pv(self):
        func = self.v1.list_persistent_volume
        self.watch_resource(func, resource.Pv)

    def watch_pvc(self):
        func = self.v1.list_namespaced_persistent_volume_claim
        if self.namespace == 'all':
            func = self.v1.list_persistent_volume_claim_for_all_namespaces
        self.watch_resource(func, resource.Pvc)

    def watch_statefulset(self):
        func = self.apps_v1.list_namespaced_stateful_set
        if self.namespace == 'all':
            func = self.apps_v1.list_stateful_set_for_all_namespaces
        self.watch_resource(func, resource.StatefulSet)

    def watch_deployments(self):
        func = self.extensions_v1beta1.list_namespaced_deployment
        if self.namespace == 'all':
            func = self.extensions_v1beta1.list_deployment_for_all_namespaces
        self.watch_resource(func, resource.Deployment)

    def watch_namespaces(self):
        self.poll_resource(self.v1.list_namespace, self.namespace_poll_time, resource.Namespace)


def parse_args():
    parser = argparse.ArgumentParser(description='Watch kube resources and keep a local cache up to date.')
    parser.add_argument('--dir', '-d', dest='dir', type=str, help='cache dir location. Default to KUBECTL_FZF_CACHE env var', default=os.environ.get('KUBECTL_FZF_CACHE', None))
    parser.add_argument("--selector", "-l", dest='selector', type=str, help='Resource selector to use', default=None)
    parser.add_argument("--node-poll-time", dest='node_poll_time', type=int, help='Time between two polls for Nodes', default=300)
    parser.add_argument("--namespace-poll-time", dest='namespace_poll_time', type=int, help='Time between two polls for Namespace', default=600)
    parser.add_argument("--namespace", "-n", dest='namespace', type=str, help='Namespace to filter. Default to current namespace. all for no filter', default=None)
    parser.add_argument("--refresh-command", dest='refresh_command', type=str, help='Command to launch when the token is expired', default=None)
    parser.add_argument("--verbose", "-v", dest='verbose', action='store_true')
    return parser.parse_args()


def start_watches(cluster, namespace, args):
    processes = []
    resource_watcher = ResourceWatcher(cluster, namespace, args)

    for f in [resource_watcher.watch_pods, resource_watcher.watch_deployments,
              resource_watcher.watch_services, resource_watcher.watch_nodes,
              resource_watcher.watch_statefulset, resource_watcher.watch_replicaset,
              resource_watcher.watch_configmap, resource_watcher.watch_endpoint,
              resource_watcher.watch_pv, resource_watcher.watch_pvc,
              resource_watcher.watch_namespaces]:
        p = multiprocessing.Process(target=f)
        p.daemon = True
        p.start()
        processes.append((p, f))
    return resource_watcher, processes


def wait_loop(resource_watcher, processes, cluster, namespace):
    while exiting is False:
        dead_processes = []
        for p, f in processes:
            p.join(1)
            if exiting is True:
                log.info('Exiting wait loop')
                return
            if not p.is_alive():
                dead_processes.append((p, f))
        for p, f in dead_processes:
            log.info('Restarting {} processes'.format(len(dead_processes)))
            processes.remove((p, f))
            # Renew login if necessary
            resource_watcher.construct_kubernetes_api()
            new_p = multiprocessing.Process(target=f)
            new_p.daemon = True
            new_p.start()
            processes.append((new_p, f))
        new_cluster, new_namespace = get_current_context()
        if cluster != new_cluster:
            log.info('Watched cluster {} != {}'.format(cluster, new_cluster))
            return
        if namespace != 'all' and namespace != new_namespace:
            log.info('Watched namespace {} != {}'.format(namespace, new_namespace))
            return


def get_current_context():
    context = config.list_kube_config_contexts()[1]['context']
    cluster = context['cluster']
    namespace = context['namespace']
    return cluster, namespace


def main():
    args = parse_args()

    signal.signal(signal.SIGINT, signal_handler)
    log_level = logging.INFO
    if args.verbose:
        log_level = logging.DEBUG
    logging.basicConfig(level=log_level, format='%(message)s')

    if not os.path.exists(args.dir):
        os.makedirs(args.dir)

    while exiting is False:
        cluster, namespace = get_current_context()
        if args.namespace is not None:
            namespace = args.namespace
        resource_watcher, processes = start_watches(cluster, namespace, args)
        wait_loop(resource_watcher, processes, cluster, namespace)
        for p, _ in processes:
            log.warn('Terminating {}'.format(p))
            p.terminate()
            p.join(1)

    log.warn('Exiting')


def stop_watches():
    log.warn('Stopping {} watches'.format(len(watches)))
    for w in watches:
        w.stop()
    del watches[:]


def signal_handler(signal, frame):
    log.warn('Signal received, closing watches')
    stop_watches()
    global exiting
    exiting = True


if __name__ == "__main__":
    main()
