#!env python2
import argparse
import logging
import os.path
import signal
import watcher
from watch_loop import WatchLoop


log = logging.getLogger('dd.' + __name__)


def parse_args():
    parser = argparse.ArgumentParser(description='Watch kube resources and keep a local cache up to date.')
    parser.add_argument('--dir', '-d', dest='dir', type=str, help='cache dir location. Default to KUBECTL_FZF_CACHE env var', default=os.environ.get('KUBECTL_FZF_CACHE', None))
    parser.add_argument("--selector", "-l", type=str, help='Resource selector to use', default=None)
    parser.add_argument("--node-poll-time", type=int, help='Time between two polls for Nodes', default=300)
    parser.add_argument("--namespace-poll-time", type=int, help='Time between two polls for Namespace', default=600)
    parser.add_argument("--namespace", "-n", type=str, help='Namespace to filter. Default to current namespace. all for no filter', default=None)
    parser.add_argument("--refresh-command", type=str, help='Command to launch when the token is expired', default=None)
    parser.add_argument("--resources-to-watch", type=str, help='List of resources to watch, separated by comma', default='Pod,Deployment,Service,Node,StatefulSet,ReplicaSet,ConfigMap,Endpoint,Pv,Pvc,Namespace')
    parser.add_argument("--verbose", "-v", action='store_true')
    return parser.parse_args()


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
        watch_loop = WatchLoop(args)
        watch_loop.start_watches()
        watch_loop.wait_loop()
        watch_loop.terminate_processes()

    log.warn('Exiting')


if __name__ == "__main__":
    main()
