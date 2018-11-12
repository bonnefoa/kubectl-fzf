from kubernetes import client, config
from kubernetes.config import kube_config
import os.path
import subprocess
import yaml
import base64
import json
import datetime
from dateutil import tz


class KubeConfiguration(object):

    def __init__(self, cluster, refresh_command):
        self.cluster = cluster
        self.refresh_command = refresh_command
        kubernetes_config = self._get_kubernetes_config()
        self.v1 = client.CoreV1Api(api_client=kubernetes_config)
        self.apps_v1 = client.AppsV1Api(api_client=kubernetes_config)
        self.extensions_v1beta1 = client.ExtensionsV1beta1Api(api_client=kubernetes_config)

    def _is_expired(self, user):
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

    def _get_kubernetes_config(self):
        res = None
        if self.refresh_command:
            filename = os.path.expanduser(kube_config.KUBE_CONFIG_DEFAULT_LOCATION)
            with open(filename) as f:
                loader = kube_config.KubeConfigLoader(
                    config_dict=yaml.load(f),
                    config_base_path=os.path.abspath(os.path.dirname(filename)))
                if self._is_expired(loader._user):
                    subprocess.call([self.refresh_command, self.cluster])

        res = config.new_client_from_config(context=self.cluster)
        return res
