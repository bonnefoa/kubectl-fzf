import unittest
import resource_watcher
import k8s_resource
import json
import logging
import sys
from kubernetes.client import api_client


logger = logging.getLogger()
logger.level = logging.DEBUG
stream_handler = logging.StreamHandler(sys.stdout)
logger.addHandler(stream_handler)


class Object(object):
    pass


class FakeResourceDumper(object):
    def __init__(self):
        self.dump_resource_called = 0
        self.number_truncated = 0

    def write_resource_to_file(self, resource, resource_dict, truncate_file):
        self.dump_resource_called = self.dump_resource_called + 1
        self.number_truncated = self.number_truncated + truncate_file
        self.resources = resource_dict


class TestResourceWatcher(unittest.TestCase):

    def setUp(self):
        args = Object()
        args.refresh_command = None
        args.node_poll_time = 10
        args.namespace_poll_time = 10
        args.dir = '/tmp/'
        args.selector = None
        self.resource_watcher = resource_watcher.ResourceWatcher('test', 'test', args)
        self.fake_resource_dumper = FakeResourceDumper()
        self.api_client = api_client.ApiClient()

    def _load_json(self, filename, resource_type, kls):
        with open(filename, 'r') as f:
            loaded_json = json.load(f)
            k8s_object = self.api_client._ApiClient__deserialize(loaded_json,
                                                                 resource_type)
            return kls(k8s_object)

    def test_same_deployment(self):
        deployment = None
        resource_dict = {}
        deployment = self._load_json('tests/deployment.json', 'V1Deployment',
                                     k8s_resource.Deployment)

        self.resource_watcher.process_resource(deployment,
                                               resource_dict,
                                               self.fake_resource_dumper)

        self.assertEqual(len(resource_dict), 1)
        self.assertEqual(len(self.fake_resource_dumper.resources), 1)
        self.assertEqual(self.fake_resource_dumper.dump_resource_called, 1)

        self.resource_watcher.process_resource(deployment,
                                               resource_dict,
                                               self.fake_resource_dumper)

        self.assertEqual(len(resource_dict), 0)
        self.assertEqual(len(self.fake_resource_dumper.resources), 0)
        self.assertEqual(self.fake_resource_dumper.number_truncated, 1)
