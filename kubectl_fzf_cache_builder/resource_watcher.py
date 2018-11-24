#!env python2
from kubernetes import watch
import logging
import os.path
import time
from urllib3.exceptions import ProtocolError, ReadTimeoutError, NewConnectionError
from configuration import KubeConfiguration


log = logging.getLogger('dd.' + __name__)
watches = []
exiting = False


class ResourceDumper(object):

    def __init__(self, dest_dir, resource_cls):
        self.resource_cls = resource_cls
        self.header = resource_cls.header()
        self.dest_file = os.path.join(dest_dir, resource_cls._dest_file())
        self.tmp_file = '{}_'.format(self.dest_file)
        self.f = open(self.dest_file, 'w')

    def write_resources_to_file(self, resources):
        log.debug('Writing {} resources in {}'.format(len(resources), self.dest_file))
        self.f.close()
        self.f = open('{}_'.format(self.dest_file), 'w')
        self.f.write('{}\n'.format(self.header))
        self.f.writelines(['{}\n'.format(str(r)) for r in resources])
        self.f.flush()
        os.rename(self.tmp_file, self.dest_file)

    def write_resource_to_file(self, resource, resource_dict, truncate_file):
        if truncate_file:
            self.write_resources_to_file(resource_dict.values())
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
        if self.namespace is not None:
            self.kube_kwargs['namespace'] = self.namespace
        self.kube_conf = None

    def check_expired_conf(self):
        self.kube_conf = KubeConfiguration(self.cluster, self.refresh_command)

    def process_resource(self, resource, resource_dict, resource_dumper):
        resource_present = resource in resource_dict
        if resource.is_deleted(resource_present):
            if resource_present:
                log.debug('Removing resource {}'.format(resource))
                resource_dict.pop(resource)
                resource_dumper.write_resource_to_file(resource, resource_dict,
                                                       True)
            return

        if not resource_present:
            log.debug('Adding resource {}'.format(resource))
            resource_dict[resource] = resource
            resource_dumper.write_resource_to_file(resource, resource_dict,
                                                   False)
            return

        if resource_dict[resource].has_changed(resource):
            log.debug('Updating resource {}'.format(resource))
            resource_dict.pop(resource)
            resource_dict[resource] = resource
            resource_dumper.write_resource_to_file(resource, resource_dict,
                                                   True)

    def _get_resource_kwargs(self, Resource):
        kwargs = self.kube_kwargs
        if not Resource._has_namespace():
            kwargs = dict(self.kube_kwargs)
            kwargs.pop('namespace', None)
        return kwargs

    def _get_namespace_str(self):
        if self.namespace is None:
            return 'all namespaces'
        return 'namespace {}'.format(self.namespace)

    def watch_resource(self, resource_cls):
        if self.kube_conf is None:
            self.kube_conf = KubeConfiguration(self.cluster, self.refresh_command)
        resource_dumper = ResourceDumper(self.dir, resource_cls)
        log.warn('Watching {} on {}, writing results in {}'.format(
            resource_cls.__name__, self._get_namespace_str(),
            resource_dumper.dest_file))
        w = watch.Watch()
        watches.append(w)
        resource_dict = {}
        kwargs = self._get_resource_kwargs(resource_cls)
        i = 0
        resource_version = 0
        list_func = resource_cls.list_func(self.kube_conf, self.namespace)
        while True:
            try:
                for resp in w.stream(list_func, resource_version=resource_version, **kwargs):
                    resource = resource_cls(resp['object'])
                    resource_version = resp['object'].metadata.resource_version
                    self.process_resource(resource, resource_dict, resource_dumper)
                    i = i + 1
                    if i % 1000 == 0:
                        log.info('Processed {} {}s'.format(i, resource_cls.__name__))
                resource_dumper.close()
            except (ReadTimeoutError, NewConnectionError, ProtocolError) as e:
                log.warn('{} watcher retrying on following error: {}'.format(resource_cls.__name__, e))
                time.sleep(1)
            except Exception as e:
                log.warn('{} watcher exiting due to {}'.format(resource_cls.__name__, e))
                time.sleep(1)
                return

    def poll_resource(self, poll_time, resource_cls):
        if self.kube_conf is None:
            self.kube_conf = KubeConfiguration(self.cluster, self.refresh_command)
        dest_file=os.path.join(self.dir, resource_cls._dest_file())
        log.info('Poll {} on {}, writing results in {}'.format(
            resource_cls.__name__, self._get_namespace_str(), dest_file))
        kwargs = self._get_resource_kwargs(resource_cls)
        resource_dumper = ResourceDumper(self.dir, resource_cls)
        list_func = resource_cls.list_func(self.kube_conf, self.namespace)
        while exiting is False:
            resp = list_func(**kwargs)
            resources = [resource_cls(item) for item in resp.items]
            resource_dumper.write_resources_to_file(resources)
            time.sleep(poll_time)
        resource_dumper.close()
        log.info('{} poll exiting {}'.format(resource_cls.__name__, exiting))


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
