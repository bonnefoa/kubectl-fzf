from kubernetes import config
import k8s_resource
import logging
import multiprocessing
import watcher


log = logging.getLogger('dd.' + __name__)


class WatchLoop():

    def __init__(self, cli_args):
        self.poll_times = {k8s_resource.Namespace: cli_args.namespace_poll_time,
                           k8s_resource.Node: cli_args.node_poll_time}
        self.resource_watcher = watcher.ResourceWatcher(self.cluster,
                                                        self.namespace, cli_args)
        self.resources_to_watch = cli_args.resources_to_watch.split(',')
        self.cluster, namespace = self._get_current_context()
        self.forced_namespace = False
        if cli_args.namespace != namespace:
            self.namespace = cli_args.namespace
            self.forced_namespace = True

    def _get_process(self, cls):
        if cls.is_poll():
            return multiprocessing.Process(target=self.resource_watcher.poll_resource,
                                           args=(self.poll_times[cls], cls))
        else:
            return multiprocessing.Process(target=self.resource_watcher.watch_resource,
                                           args=(cls,))

    def _get_current_context(self):
        context = config.list_kube_config_contexts()[1]['context']
        cluster = context['cluster']
        namespace = context['namespace']
        return cluster, namespace

    def _resource_for_name(self, class_name):
        m = __import__('k8s_resource', globals(), locals(), class_name)
        c = getattr(m, class_name)
        return c

    def start_watches(self):
        self.processes = []
        for r in self.resources_to_watch:
            cls = self._resource_for_name(r)
            p = None
            p = self._get_process(cls)
            p.daemon = True
            p.start()
            self.processes.append((p, cls))

    def wait_loop(self):
        while watcher.exiting is False:
            dead_processes = []
            for p, cls in self.processes:
                p.join(1)
                if watcher.exiting is True:
                    log.info('Exiting wait loop')
                    return
                if not p.is_alive():
                    dead_processes.append((p, cls))
            for p, cls in dead_processes:
                log.info('Restarting {} processes'.format(len(dead_processes)))
                self.processes.remove((p, cls))
                # Renew login if necessary
                self.resource_watcher.check_expired_conf()
                new_p = self._get_process(cls)
                new_p.daemon = True
                new_p.start()
                self.processes.append((new_p, cls))
            new_cluster, new_namespace = self.get_current_context()
            if self.cluster != new_cluster:
                log.info('Watched cluster {} != {}'.format(self.cluster,
                                                           new_cluster))
                return
            if self.forced_namespace is False and self.namespace != new_namespace:
                log.info('Watched namespace {} != {}'.format(self.namespace,
                                                             new_namespace))
                return

    def terminate_processes(self):
        for p, _ in self.processes:
            log.warn('Terminating {}'.format(p))
            p.terminate()
            p.join(1)
