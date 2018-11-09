import logging
from datetime import datetime


EXCLUDED_LABELS = ['pod-template-generation', 'app.kubernetes.io/name', 'controller-revision-hash',
                   'app.kubernetes.io/managed-by', 'pod-template-hash', 'statefulset.kubernetes.io/pod-name',
                   'controler-uid']
log = logging.getLogger('dd.' + __name__)


class Resource(object):

    def __init__(self, resource):
        self.name = resource.metadata.name
        if hasattr(resource.metadata, 'namespace'):
            self.namespace = resource.metadata.namespace
        self.labels = resource.metadata.labels or {}
        for l in EXCLUDED_LABELS:
            self.labels.pop(l, None)
        self.label_keys = self.labels.keys()
        self.label_keys.sort()
        if hasattr(resource, 'status') and hasattr(resource.status, 'start_time'):
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
            return ','.join(['{}={}'.format(k, self.labels[k]) for k in self.label_keys])
        else:
            return 'None'

    def _list_str(self, lst):
        if lst:
            return ','.join(lst)
        else:
            return 'None'

    def _selector_str(self):
        if self.selector:
            return ','.join(self.selector)
        else:
            return 'None'

    @classmethod
    def _dest_file(cls):
        return '{}s'.format(cls.__name__).lower()

    @staticmethod
    def _has_namespace():
        return True

    def __hash__(self):
        return hash((self.name, self.namespace))

    def __eq__(self, other):
        return self.name == other.name and self.namespace == other.namespace


class Pod(Resource):

    def __init__(self, pod):
        Resource.__init__(self, pod)
        self.host_ip = pod.status.host_ip
        self.node_name = pod.spec.node_name
        self.phase = self._get_phase(pod)

    def _get_phase(self, pod):
        if pod.status.container_statuses:
            for s in pod.status.container_statuses:
                if s.state.waiting:
                    if 'Completed' not in s.state.waiting.reason:
                        return s.state.waiting.reason
        return pod.status.phase

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

    @staticmethod
    def header():
        return "Namespace Name Labels HostIp NodeName Phase Age"


class ReplicaSet(Resource):

    def __init__(self, sts):
        Resource.__init__(self, sts)
        self.replicas = sts.status.replicas or 0
        self.ready_replicas = sts.status.ready_replicas or 0
        self.available_replicas = sts.status.available_replicas or 0
        self.selector = ['{}={}'.format(k, v)
                         for k, v in sts.spec.selector.match_labels.items()
                         if k not in EXCLUDED_LABELS]

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(str(self.replicas))
        content.append(str(self.available_replicas))
        content.append(str(self.ready_replicas))
        content.append(self._selector_str())
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def header():
        return "Namespace Name Labels Replicas AvailableReplicas ReadyReplicas Selector Age"


class ConfigMap(Resource):

    def __init__(self, config_map):
        Resource.__init__(self, config_map)

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def header():
        return "Namespace Name Labels Age"


class StatefulSet(Resource):

    def __init__(self, sts):
        Resource.__init__(self, sts)
        self.selector = ['{}={}'.format(k, v)
                         for k, v in sts.spec.selector.match_labels.items()
                         if k not in EXCLUDED_LABELS]
        self.current_replicas = sts.status.current_replicas or sts.status.ready_replicas
        self.replicas = sts.status.replicas

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append('{}/{}'.format(self.current_replicas, self.replicas))
        content.append(self._selector_str())
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def header():
        return "Namespace Name Labels Replicas Selector Age"


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

    @staticmethod
    def header():
        return "Namespace Name Labels Age"


class Endpoint(Resource):

    def __init__(self, endpoint):
        Resource.__init__(self, endpoint)
        self.ready_ips = []
        self.ready_pods = []
        self.not_ready_ips = []
        self.not_ready_pods = []
        self._fill_ips(endpoint)

    def _fill_ips(self, endpoint):
        if endpoint.subsets is None:
            return
        for subset in endpoint.subsets:
            if subset.addresses:
                for add in subset.addresses:
                    self.ready_ips.append(add.ip)
                    target = add.target_ref
                    if target and target.kind == "Pod":
                        self.ready_pods.append(target.name)

            if subset.not_ready_addresses:
                for add in subset.not_ready_addresses:
                    self.not_ready_ips.append(add.ip)
                    target = add.target_ref
                    if target and target.kind == "Pod":
                        self.not_ready_pods.append(target.name)

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(self._resource_age())
        content.append(self._list_str(self.ready_ips))
        content.append(self._list_str(self.ready_pods))
        content.append(self._list_str(self.not_ready_ips))
        content.append(self._list_str(self.not_ready_pods))
        return ' '.join(content)

    @staticmethod
    def header():
        return "Namespace Name Labels Age ReadyIps ReadyPods NotReadyIps NotReadyPods"


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
        content.append(self._list_str(self.roles))
        content.append(self.instance_type)
        content.append(self.zone)
        content.append(self.internal_ip)
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def _has_namespace():
        return False

    @staticmethod
    def header():
        return "Name Labels Roles InstanceType Zone InternalIp Age"


class Service(Resource):

    def __init__(self, service):
        Resource.__init__(self, service)
        self.type = service.spec.type
        self.cluster_ip = service.spec.cluster_ip
        self.ports = []
        self.selector = []
        if service.spec.ports:
            self.ports = ['{}:{}'.format(p.name, p.port)
                          for p in service.spec.ports]
        if service.spec.selector is not None:
            self.selector = ['{}={}'.format(k, v)
                             for k, v in service.spec.selector.items()
                             if k not in EXCLUDED_LABELS]

    def __str__(self):
        content = []
        content.append(self.namespace)
        content.append(self.name)
        content.append(self._label_str())
        content.append(self.type)
        content.append(self.cluster_ip)
        content.append(self._list_str(self.ports))
        content.append(self._selector_str())
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def header():
        return "Namespace Name Labels Type ClusterIp Ports Selector Age"


class Namespace(Resource):

    def __init__(self, namespace):
        Resource.__init__(self, namespace)

    def __str__(self):
        content = []
        content.append(self.name)
        content.append(self._label_str())
        content.append(self._resource_age())
        return ' '.join(content)

    @staticmethod
    def header():
        return "Name Labels Age"

    @staticmethod
    def _has_namespace():
        return False
