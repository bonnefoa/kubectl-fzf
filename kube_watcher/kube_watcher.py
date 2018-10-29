#!env python2
import argparse
from kubernetes import client, config, watch
import logging
import os.path
import threading
import signal


log = logging.getLogger('dd.' + __name__)
EXCLUDED_LABELS=['pod-template-generation', 'app.kubernetes.io/name', 'controller-revision-hash',
                 'app.kubernetes.io/managed-by', 'pod-template-hash', 'statefulset.kubernetes.io/pod-name',
                 'controler-uid']

watches = []


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
        content.append(','.join(['{}={}'.format(k, v) for k, v in self.labels.items()]))
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

    def __init__(self, v1, extensions_v1beta1, args):
        self.pods = set()
        self.deployments = set()
        self.namespace = args.namespace
        self.v1 = v1
        self.extensions_v1beta1 = extensions_v1beta1
        self.pod_file_path = os.path.join(args.dir, 'pods')
        self.deployment_file_path = os.path.join(args.dir, 'deployments')
        self.pod_file = open(self.pod_file_path, 'w')
        self.deployment_file = open(self.deployment_file_path, 'w')

        self.kube_kwargs = {'_request_timeout': 600}
        if args.selector:
            self.kube_kwargs['label_selector'] = args.selector
        if args.namespace != 'all':
            self.kube_kwargs['namespace'] = args.namespace

    def write_resource_to_file(self, resource, resources, resource_was_present, f):
        if resource_was_present:
            log.debug('Truncating as resourse {} is already present'.format(resource))
            f.seek(0)
            f.truncate()
            f.writelines(['{}\n'.format(p) for p in resources])
        else:
            f.write('{}\n'.format(str(resource)))
        f.flush()

    def watch_pods(self):
        w = watch.Watch()
        watches.append(w)
        func = self.v1.list_namespaced_pod
        if self.namespace == 'all':
            func = self.v1.list_pod_for_all_namespaces
        log.info('Watching pods on namespace {}, writing results in {}'.format(
            self.namespace, self.pod_file_path))
        for p in w.stream(func, **self.kube_kwargs):
            pod = Pod(p['object'])
            pod_was_present = False
            if pod in self.pods:
                pod_was_present = True
            self.pods.add(pod)
            self.write_resource_to_file(pod, self.pods, pod_was_present, self.pod_file)
        self.pod_file.close()

    def watch_deployments(self):
        w = watch.Watch()
        watches.append(w)
        func = self.extensions_v1beta1.list_namespaced_deployment
        if self.namespace == 'all':
            func = self.extensions_v1beta1.list_deployment_for_all_namespaces
        log.info('Watching deployments on namespace {}, writing results in {}'.format(
            self.namespace, self.deployment_file_path))
        for p in w.stream(func, **self.kube_kwargs):
            deployment = Deployment(p['object'])
            deployment_was_present = False
            if deployment in self.deployments:
                deployment_was_present = True
            self.deployments.add(deployment)
            self.write_resource_to_file(deployment, self.deployments, deployment_was_present, self.deployment_file)
        self.deployment_file.close()


def parse_args():
    parser = argparse.ArgumentParser(description='Watch kube resources and keep a local cache up to date.')
    parser.add_argument('--dir', '-d', dest='dir', type=str, help='cache dir location', default=os.environ.get('KUBECTL_FZF_CACHE', None))
    parser.add_argument("--selector", "-l", dest='selector', type=str, help='Resource selector to use', default=None)
    parser.add_argument("--namespace", "-n", dest='namespace', type=str, help='Namespace to filter. Default to current namespace. all for no filter', default=None)
    parser.add_argument("--verbose", "-v", dest='verbose', action='store_true')
    return parser.parse_args()


def main():
    args = parse_args()

    signal.signal(signal.SIGINT, signal_handler)
    log_level = logging.INFO
    if args.verbose:
        log_level = logging.DEBUG
    logging.basicConfig(level=log_level, format='%(message)s')

    config.load_kube_config()
    current_namespace = config.list_kube_config_contexts()[1]['context']['namespace']
    if args.namespace is None:
        args.namespace = current_namespace
    v1 = client.CoreV1Api()
    extensions_v1beta1 = client.ExtensionsV1beta1Api()
    if not os.path.exists(args.dir):
        os.makedirs(args.dir)
    resource_watcher = ResourceWatcher(v1, extensions_v1beta1, args)
    t_pods = threading.Thread(target=resource_watcher.watch_pods)
    t_pods.daemon = True
    t_pods.start()

    t_deployments = threading.Thread(target=resource_watcher.watch_deployments)
    t_deployments.daemon = True
    t_deployments.start()

    while True:
        t_deployments.join(1000)
        if not t_pods.isAlive():
            break
        if not t_deployments.isAlive():
            break


def signal_handler(signal, frame):
    log.warn('Signal received, existing')
    for w in watches:
        w.stop()


if __name__ == "__main__":
    main()
