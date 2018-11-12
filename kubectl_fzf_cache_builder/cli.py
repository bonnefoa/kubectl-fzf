#!env python2
import argparse
from kubernetes import config
import k8s_resource
import logging
import os.path
import multiprocessing
import signal
import watcher


log = logging.getLogger('dd.' + __name__)


def parse_args():
    parser = argparse.ArgumentParser(description='Watch kube resources and keep a local cache up to date.')
    parser.add_argument('--dir', '-d', dest='dir', type=str, help='cache dir location. Default to KUBECTL_FZF_CACHE env var', default=os.environ.get('KUBECTL_FZF_CACHE', None))
    parser.add_argument("--selector", "-l", type=str, help='Resource selector to use', default=None)
    parser.add_argument("--node-poll-time", type=int, help='Time between two polls for Nodes', default=300)
    parser.add_argument("--namespace-poll-time", type=int, help='Time between two polls for Namespace', default=600)
    parser.add_argument("--namespace", "-n", type=str, help='Namespace to filter. Default to current namespace. all for no filter', default=None)
    parser.add_argument("--refresh-command", type=str, help='Command to launch when the token is expired', default=None)
    parser.add_argument("--resource-to-watch", type=str, help='List of resources to watch, separated by comma', default='Pod,Deployment,Service,Node,StatefulSet,ReplicaSet,ConfigMap,Endpoint,Pv,Pvc,Namespace')
    parser.add_argument("--verbose", "-v", action='store_true')
    return parser.parse_args()


def get_poll_times(args):
    return {k8s_resource.Namespace: args.namespace_poll_time,
            k8s_resource.Node: args.node_poll_time}


def get_process(cls, resource_watcher, poll_times):
    if cls.is_poll():
        return multiprocessing.Process(target=resource_watcher.poll_resource,
                                       args=(poll_times[cls], cls))
    else:
        return multiprocessing.Process(target=resource_watcher.watch_resource,
                                       args=(cls,))


def class_for_name(module_name, class_name):
    m = __import__(module_name, globals(), locals(), class_name)
    c = getattr(m, class_name)
    return c


def start_watches(cluster, namespace, poll_times, args):
    processes = []
    resource_watcher = watcher.ResourceWatcher(cluster, namespace, args)

    for r in args.resource_to_watch.split(','):
        cls = class_for_name('k8s_resource', r)
        p = None
        p = get_process(cls, resource_watcher, poll_times)
        p.daemon = True
        p.start()
        processes.append((p, cls))
    return resource_watcher, processes


def wait_loop(resource_watcher, processes, cluster, namespace, poll_times):
    while watcher.exiting is False:
        dead_processes = []
        for p, cls in processes:
            p.join(1)
            if watcher.exiting is True:
                log.info('Exiting wait loop')
                return
            if not p.is_alive():
                dead_processes.append((p, cls))
        for p, cls in dead_processes:
            log.info('Restarting {} processes'.format(len(dead_processes)))
            processes.remove((p, cls))
            # Renew login if necessary
            resource_watcher.check_expired_conf()
            new_p = get_process(cls, resource_watcher, poll_times)
            new_p.daemon = True
            new_p.start()
            processes.append((new_p, cls))
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

    signal.signal(signal.SIGINT, watcher.signal_handler)
    log_level = logging.INFO
    if args.verbose:
        log_level = logging.DEBUG
    logging.basicConfig(level=log_level, format='%(message)s')

    if not os.path.exists(args.dir):
        os.makedirs(args.dir)

    while watcher.exiting is False:
        cluster, namespace = get_current_context()
        if args.namespace is not None:
            namespace = args.namespace
        poll_times = get_poll_times(args)
        resource_watcher, processes = start_watches(cluster, namespace, poll_times, args)
        wait_loop(resource_watcher, processes, cluster, namespace, poll_times)
        for p, _ in processes:
            log.warn('Terminating {}'.format(p))
            p.terminate()
            p.join(1)

    log.warn('Exiting')


if __name__ == "__main__":
    main()
