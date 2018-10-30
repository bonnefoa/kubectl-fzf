#!env python2
import argparse
from kubernetes import client, config, watch
import logging
import os.path
import multiprocessing
import signal


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
        if self.labels:
            content.append(','.join(['{}={}'.format(k, v) for k, v in self.labels.items()]))
        else:
            content.append('None')
        content.append(str(self.host_ip))
        content.append(str(self.node_name))
        content.append(self.phase)
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


class ResourceWatcher(object):

    def __init__(self, cluster, namespace, args):
        self.cluster = cluster
        self.namespace = namespace
        self.v1 = client.CoreV1Api()
        self.extensions_v1beta1 = client.ExtensionsV1beta1Api()
        self.pod_file_path = os.path.join(args.dir, 'pods')
        self.deployment_file_path = os.path.join(args.dir, 'deployments')

        self.kube_kwargs = {'_request_timeout': 600}
        if args.selector:
            self.kube_kwargs['label_selector'] = args.selector
        if self.namespace != 'all':
            self.kube_kwargs['namespace'] = self.namespace

    def write_resource_to_file(self, resource, resources, resource_was_present, f):
        if resource_was_present:
            log.debug('Truncating file since resource {} is already present'.format(resource))
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
                resources.add(resource)
                self.write_resource_to_file(resource, resources, resource_was_present, dest)
        log.info('{} watcher exiting'.format(Resource.__name__))

    def watch_pods(self):
        func = self.v1.list_namespaced_pod
        if self.namespace == 'all':
            func = self.v1.list_pod_for_all_namespaces
        self.watch_resource(func, Pod, self.pod_file_path)

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
    t_pods = multiprocessing.Process(target=resource_watcher.watch_pods)
    t_pods.daemon = True
    t_pods.start()
    processes.append(t_pods)

    t_deployments = multiprocessing.Process(target=resource_watcher.watch_deployments)
    t_deployments.daemon = True
    t_deployments.start()
    processes.append(t_deployments)

    return processes


def wait_loop(processes, cluster, namespace):
    while True:
        for p in processes:
            p.join(1)
            if not p.is_alive():
                log.info('A process has joined, exiting wait loop')
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

    config.load_kube_config()

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
