#!env python2
import argparse
from kubernetes import client, config, watch
import logging
import os.path
import multiprocessing
import signal
from datetime import datetime
import subprocess
import time


log = logging.getLogger('dd.' + __name__)
EXCLUDED_LABELS=['pod-template-generation', 'app.kubernetes.io/name', 'controller-revision-hash',
                 'app.kubernetes.io/managed-by', 'pod-template-hash', 'statefulset.kubernetes.io/pod-name',
                 'controler-uid']

watches = []
exiting = False


def get_kubernetes_config(cluster, refresh_command):
    res = None
    try:
        res = config.new_client_from_config(context=cluster)
    except config.config_exception.ConfigException:
        if refresh_command:
            subprocess.call([refresh_command, cluster])
            res = config.new_client_from_config(context=cluster)
    return res


class Resource(object):

    def __init__(self, resource):
        self.name = resource.metadata.name
        self.namespace = resource.metadata.namespace
        self.labels = resource.metadata.labels or {}
        for l in EXCLUDED_LABELS:
            self.labels.pop(l, None)
        if hasattr(resource.status, 'start_time'):
            self.start_time = resource.status.start_time
        else:
            self.start_time = resource.metadata.creation_timestamp
        self.is_deleted = resource.metadata.deletion_timestamp is not None

    def _resource_age(self):
        if self.start_time:
            s = (datetime.now(self.start_time.tzinfo) - self.start_time).total_seconds()
            days, remainder = divmod(s, 86400)
            hours, remainder = divmod(s, 3600)
            minutes, _ = divmod(remainder, 60)
            if days:
                return '{}d'.format(int(days))
            elif hours:
                return '{}h'.format(int(hours))
            else:
                return '{}m'.format(int(minutes))
        else:
            return 'None'

    def _label_str(self):
        if self.labels:
            return ','.join(['{}={}'.format(k, v) for k, v in self.labels.items()])
        else:
            return 'None'

    @classmethod
    def _dest_file(cls):
        return '{}s'.format(cls.__name__).lower()

    @staticmethod
    def _has_namespace():
        return True

    def __hash__(self):
        return hash(self.name)

    def __eq__(self, other):
        return self.name == other.name


class Pod(Resource):

    def __init__(self, pod):
        Resource.__init__(self, pod)
        self.host_ip = pod.status.host_ip
        self.node_name = pod.spec.node_name
        self.phase = pod.status.phase
        if self._is_clb(pod):
            self.phase = 'CrashLoopBackoff'

    def _is_clb(self, pod):
        if pod.status.container_statuses is None:
            return False
        for s in pod.status.container_statuses:
            if s.state.waiting:
                if 'CrashLoopBackOff' == s.state.waiting.reason:
                    return True
        return False

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(str(self.host_ip))
        content.append(str(self.node_name))
        content.append(self.phase)
        content.append(self._resource_age())
        return ' '.join(content)


class Deployment(Resource):

    def __init__(self, deployment):
        Resource.__init__(self, deployment)

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(self._resource_age())
        return ' '.join(content)


class Node(Resource):

    def __init__(self, node):
        Resource.__init__(self, node)
        self.roles = []
        for k in self.labels:
            if k.startswith('node-role.kubernetes.io/'):
                self.roles.append(k.split('/')[1])
        self.instance_type = self.labels.get('beta.kubernetes.io/instance-type',
                                             'None')
        self.zone = self.labels.get('failure-domain.beta.kubernetes.io/zone',
                                    'None')
        self.internal_ip = 'None'
        for address in node.status.addresses:
            if address.type == 'InternalIP':
                self.internal_ip = address.address

    def __str__(self):
        content = []
        content.append(self.name)
        content.append(self._label_str())
        content.append(','.join(self.roles))
        content.append(self.instance_type)
        content.append(self.zone)
        content.append(self.internal_ip)
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def _has_namespace():
        return False


class Service(Resource):

    def __init__(self, service):
        Resource.__init__(self, service)
        self.type = service.spec.type
        self.cluster_ip = service.spec.cluster_ip
        self.ports = []
        if service.spec.ports:
            self.ports = ['{}:{}'.format(p.name, p.port)
                          for p in service.spec.ports]
        self.selector = ['{}={}'.format(k, v)
                         for k, v in service.spec.selector.items()]

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(self.type)
        content.append(self.cluster_ip)
        if self.ports:
            content.append(','.join(self.ports))
        else:
            content.append('None')
        if self.selector:
            content.append(','.join(self.selector))
        else:
            content.append('None')
        content.append(self._resource_age())
        return ' '.join(content)


class ResourceWatcher(object):

    def __init__(self, cluster, namespace, args):
        self.cluster = cluster
        self.namespace = namespace
        kubernetes_config = get_kubernetes_config(cluster, args.refresh_command)
        self.v1 = client.CoreV1Api(api_client=kubernetes_config)
        self.extensions_v1beta1 = client.ExtensionsV1beta1Api(api_client=kubernetes_config)
        self.poll_time = args.poll_time
        self.kube_kwargs = {'_request_timeout': 600}
        if args.selector:
            self.kube_kwargs['label_selector'] = args.selector
        if self.namespace != 'all':
            self.kube_kwargs['namespace'] = self.namespace

    def write_resource_to_file(self, resource, resources, truncate_file, f):
        if truncate_file:
            log.debug('Truncating file {}'.format(resource))
            f.seek(0)
            f.truncate()
            f.writelines(['{}\n'.format(r) for r in resources])
        else:
            f.write('{}\n'.format(str(resource)))
        f.flush()

    def write_resources_to_file(self, resources, f):
        f.writelines(['{}\n'.format(r) for r in resources])
        f.flush()

    def process_resource(self, resource, resources, dest):
        do_truncate = False
        if resource.is_deleted and resource in resources:
            log.info('Removing resource {}'.format(resource))
            resources.remove(resource)
            do_truncate = True
        else:
            if resource in resources:
                log.info('Updating resource {}'.format(resource))
                do_truncate = True
                resources.remove(resource)
            resources.add(resource)
        self.write_resource_to_file(resource, resources, do_truncate, dest)

    def _get_resource_kwargs(self, Resource):
        kwargs = self.kube_kwargs
        if not Resource._has_namespace():
            kwargs = dict(self.kube_kwargs)
            kwargs.pop('namespace', None)
        return kwargs

    def watch_resource(self, func, ResourceCls):
        dest_file=ResourceCls._dest_file()
        log.info('Watching {} on namespace {}, writing results in {}'.format(
            ResourceCls.__name__, self.namespace, dest_file))
        w = watch.Watch()
        watches.append(w)
        resources = set()
        with open(dest_file, 'w') as dest:
            kwargs = self._get_resource_kwargs(ResourceCls)
            for resp in w.stream(func, **kwargs):
                resource = ResourceCls(resp['object'])
                self.process_resource(resource, resources, dest)
        log.info('{} watcher exiting'.format(ResourceCls.__name__))

    def poll_resource(self, func, ResourceCls):
        dest_file=ResourceCls._dest_file()
        log.info('Poll {} on namespace {}, writing results in {}'.format(
            ResourceCls.__name__, self.namespace, dest_file))
        kwargs = self._get_resource_kwargs(ResourceCls)
        while exiting is False:
            resp = func(**kwargs)
            resources = [ResourceCls(item) for item in resp.items]
            with open(dest_file, 'w') as dest:
                self.write_resources_to_file(resources, dest)
            time.sleep(self.poll_time)
        log.info('{} poll exiting {}'.format(ResourceCls.__name__, exiting))

    def watch_pods(self):
        func = self.v1.list_namespaced_pod
        if self.namespace == 'all':
            func = self.v1.list_pod_for_all_namespaces
        self.watch_resource(func, Pod)

    def watch_services(self):
        func = self.v1.list_namespaced_service
        if self.namespace == 'all':
            func = self.v1.list_service_for_all_namespaces
        self.watch_resource(func, Service)

    def watch_nodes(self):
        self.poll_resource(self.v1.list_node, Node)

    def watch_deployments(self):
        func = self.extensions_v1beta1.list_namespaced_deployment
        if self.namespace == 'all':
            func = self.extensions_v1beta1.list_deployment_for_all_namespaces
        self.watch_resource(func, Deployment)


def parse_args():
    parser = argparse.ArgumentParser(description='Watch kube resources and keep a local cache up to date.')
    parser.add_argument('--dir', '-d', dest='dir', type=str, help='cache dir location. Default to KUBECTL_FZF_CACHE env var', default=os.environ.get('KUBECTL_FZF_CACHE', None))
    parser.add_argument("--selector", "-l", dest='selector', type=str, help='Resource selector to use', default=None)
    parser.add_argument("--poll-time", dest='poll_time', type=int, help='Time between two list requests for polled resources', default=300)
    parser.add_argument("--namespace", "-n", dest='namespace', type=str, help='Namespace to filter. Default to current namespace. all for no filter', default=None)
    parser.add_argument("--refresh-command", dest='refresh_command', type=str, help='Command to launch when the token is expired', default=None)
    parser.add_argument("--verbose", "-v", dest='verbose', action='store_true')
    return parser.parse_args()


def start_watches(cluster, namespace, args):
    processes = []
    resource_watcher = ResourceWatcher(cluster, namespace, args)
    for f in [resource_watcher.watch_pods, resource_watcher.watch_deployments,
              resource_watcher.watch_services, resource_watcher.watch_nodes]:
        p = multiprocessing.Process(target=f)
        p.daemon = True
        p.start()
        processes.append(p)
    return processes


def wait_loop(processes, cluster, namespace):
    while exiting is False:
        for p in processes:
            p.join(1)
            if not p.is_alive():
                log.info('A process is dead, exiting wait loop')
                return
        new_cluster, new_namespace = get_current_context()
        if cluster != new_cluster:
            log.info('Watched cluster {} != {}'.format(cluster, new_cluster))
            return
        if namespace != new_namespace:
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
        processes = start_watches(cluster, namespace, args)
        wait_loop(processes, cluster, namespace)
        for p in processes:
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
