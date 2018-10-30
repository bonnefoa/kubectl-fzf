#!env python2
import argparse
from kubernetes import client, config, watch
import logging
import os.path
import multiprocessing
import signal
from datetime import datetime


log = logging.getLogger('dd.' + __name__)
EXCLUDED_LABELS=['pod-template-generation', 'app.kubernetes.io/name', 'controller-revision-hash',
                 'app.kubernetes.io/managed-by', 'pod-template-hash', 'statefulset.kubernetes.io/pod-name',
                 'controler-uid']

watches = []
retry = True


class Pod(object):

    def __init__(self, pod):
        self.name = pod.metadata.name
        self.namespace = pod.metadata.namespace
        self.labels = pod.metadata.labels or {}
        for l in EXCLUDED_LABELS:
            self.labels.pop(l, None)
        self.host_ip = pod.status.host_ip
        self.node_name = pod.spec.node_name
        self.phase = pod.status.phase
        self.start_time = pod.status.start_time
        if self._is_clb(pod):
            self.phase = 'CrashLoopBackoff'
        self.is_deleted = False
        if pod.status.container_statuses is not None:
            self.is_deleted = all([s.state.terminated for s in pod.status.container_statuses])

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
        if self.labels:
            content.append(','.join(['{}={}'.format(k, v) for k, v in self.labels.items()]))
        else:
            content.append('None')
        content.append(str(self.host_ip))
        content.append(str(self.node_name))
        content.append(self.phase)

        if self.start_time:
            s = (datetime.now(self.start_time.tzinfo) - self.start_time).total_seconds()
            days, remainder = divmod(s, 86400)
            hours, remainder = divmod(s, 3600)
            minutes, _ = divmod(remainder, 60)
            if days:
                content.append('{}d'.format(int(days)))
            elif hours:
                content.append('{}h'.format(int(hours)))
            else:
                content.append('{}m'.format(int(minutes)))
        else:
            content.append('None')
        return ' '.join(content)

    def __hash__(self):
        return hash(self.name)

    def __eq__(self, other):
        return self.name == other.name


class Deployment(object):

    def __init__(self, deployment):
        self.name = deployment.metadata.name
        self.namespace = deployment.metadata.namespace
        self.labels = deployment.metadata.labels or {}
        self.is_deleted = False
        for l in EXCLUDED_LABELS:
            self.labels.pop(l, None)

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(','.join(['{}={}'.format(k, v) for k, v in self.labels.items()]))
        return ' '.join(content)

    def __hash__(self):
        return hash(self.name)

    def __eq__(self, other):
        return self.name == other.name


class Service(object):

    def __init__(self, deployment):
        self.name = deployment.metadata.name
        self.namespace = deployment.metadata.namespace
        self.labels = deployment.metadata.labels or {}
        self.type = deployment.spec.type
        self.cluster_ip = deployment.spec.cluster_ip
        self.ports = []
        if deployment.spec.ports:
            self.ports = ['{}:{}'.format(p.name, p.port) for p in deployment.spec.ports]
        self.is_deleted = False
        for l in EXCLUDED_LABELS:
            self.labels.pop(l, None)

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self.type)
        content.append(self.cluster_ip)
        content.append(','.join(self.ports))
        return ' '.join(content)

    def __hash__(self):
        return hash(self.name)

    def __eq__(self, other):
        return self.name == other.name


class ResourceWatcher(object):

    def __init__(self, cluster, namespace, args):
        self.cluster = cluster
        self.namespace = namespace
        self.v1 = client.CoreV1Api(
            api_client=config.new_client_from_config(context=cluster))
        self.extensions_v1beta1 = client.ExtensionsV1beta1Api(
            api_client=config.new_client_from_config(context=cluster))
        self.pod_file_path = os.path.join(args.dir, 'pods')
        self.service_file_path = os.path.join(args.dir, 'services')
        self.deployment_file_path = os.path.join(args.dir, 'deployments')

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
            f.writelines(['{}\n'.format(p) for p in resources])
        else:
            f.write('{}\n'.format(str(resource)))
        f.flush()

    def watch_resource(self, func, Resource, dest_file):
        log.info('Watching {} on namespace {}, writing results in {}'.format(
            Resource.__name__, self.namespace, dest_file))
        w = watch.Watch()
        watches.append(w)
        resources = set()
        with open(dest_file, 'w') as dest:
            for p in w.stream(func, **self.kube_kwargs):
                resource = Resource(p['object'])
                resource_was_present = False
                if resource in resources:
                    resource_was_present = True
                    resources.remove(resource)
                if resource.is_deleted and resource in resources:
                    resources.remove(resource)
                else:
                    resources.add(resource)
                self.write_resource_to_file(resource, resources, resource_was_present, dest)
        log.info('{} watcher exiting'.format(Resource.__name__))

    def watch_pods(self):
        func = self.v1.list_namespaced_pod
        if self.namespace == 'all':
            func = self.v1.list_pod_for_all_namespaces
        self.watch_resource(func, Pod, self.pod_file_path)

    def watch_services(self):
        func = self.v1.list_namespaced_service
        if self.namespace == 'all':
            func = self.v1.list_service_for_all_namespaces
        self.watch_resource(func, Service, self.service_file_path)

    def watch_deployments(self):
        func = self.extensions_v1beta1.list_namespaced_deployment
        if self.namespace == 'all':
            func = self.extensions_v1beta1.list_deployment_for_all_namespaces
        self.watch_resource(func, Deployment, self.deployment_file_path)


def parse_args():
    parser = argparse.ArgumentParser(description='Watch kube resources and keep a local cache up to date.')
    parser.add_argument('--dir', '-d', dest='dir', type=str, help='cache dir location', default=os.environ.get('KUBECTL_FZF_CACHE', None))
    parser.add_argument("--selector", "-l", dest='selector', type=str, help='Resource selector to use', default=None)
    parser.add_argument("--namespace", "-n", dest='namespace', type=str, help='Namespace to filter. Default to current namespace. all for no filter', default=None)
    parser.add_argument("--verbose", "-v", dest='verbose', action='store_true')
    return parser.parse_args()


def start_watches(cluster, namespace, args):
    processes = []
    resource_watcher = ResourceWatcher(cluster, namespace, args)
    for f in [resource_watcher.watch_pods, resource_watcher.watch_deployments, resource_watcher.watch_services]:
        p = multiprocessing.Process(target=f)
        p.daemon = True
        p.start()
        processes.append(p)
    return processes


def wait_loop(processes, cluster, namespace):
    while retry:
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

    while retry:
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
    global retry
    retry = False


if __name__ == "__main__":
    main()
