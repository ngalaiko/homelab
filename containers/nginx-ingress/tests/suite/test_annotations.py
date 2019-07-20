import pytest
import yaml
from kubernetes.client import ExtensionsV1beta1Api

from suite.fixtures import PublicEndpoint
from suite.resources_utils import ensure_connection_to_public_endpoint, \
    get_ingress_nginx_template_conf, \
    get_first_pod_name, create_example_app, wait_until_all_pods_are_ready, \
    delete_common_app, create_items_from_yaml, delete_items_from_yaml, \
    wait_before_test, replace_configmap_from_yaml, get_events
from suite.yaml_utils import get_first_ingress_host_from_yaml, get_names_from_yaml
from settings import TEST_DATA, DEPLOYMENTS


def get_event_count(event_text, events_list) -> int:
    for i in range(len(events_list) - 1, -1, -1):
        if event_text in events_list[i].message:
            return events_list[i].count
    return 0


def assert_event_count_increased(event_text, count, events_list):
    for i in range(len(events_list) - 1, -1, -1):
        if event_text in events_list[i].message:
            assert events_list[i].count > count
            return
    pytest.fail(f"Failed to find the event \"{event_text}\" in the list. Exiting...")


def generate_ingresses_with_annotation(yaml_manifest, annotations) -> []:
    """
    Generate an Ingress item with an annotation.

    :param yaml_manifest: an absolute path to a file
    :param annotations:
    :return: []
    """
    res = []
    with open(yaml_manifest) as f:
        docs = yaml.load_all(f)
        for doc in docs:
            if doc['kind'] == 'Ingress':
                doc['metadata']['annotations'].update(annotations)
                res.append(doc)
    return res


def replace_ingress(extensions_v1_beta1: ExtensionsV1beta1Api, name, namespace, body) -> str:
    """
    Replace an Ingress based on a dict.

    :param extensions_v1_beta1: ExtensionsV1beta1Api
    :param name:
    :param namespace: namespace
    :param body: dict
    :return: str
    """
    print(f"Replace a Ingress: {name}")
    resp = extensions_v1_beta1.replace_namespaced_ingress(name, namespace, body)
    print(f"Ingress replaced with name '{name}'")
    return resp.metadata.name


def replace_ingresses_from_yaml(extensions_v1_beta1: ExtensionsV1beta1Api, namespace, yaml_manifest) -> None:
    """
    Parse file and replace all Ingresses based on its contents.

    :param extensions_v1_beta1: ExtensionsV1beta1Api
    :param namespace: namespace
    :param yaml_manifest: an absolute path to a file
    :return:
    """
    print(f"Replace an Ingresses from yaml")
    with open(yaml_manifest) as f:
        docs = yaml.load_all(f)
        for doc in docs:
            if doc['kind'] == 'Ingress':
                replace_ingress(extensions_v1_beta1, doc['metadata']['name'], namespace, doc)


class AnnotationsSetup:
    """Encapsulate Annotations example details.

    Attributes:
        public_endpoint: PublicEndpoint
        ingress_name:
        ingress_pod_name:
        ingress_host:
        namespace: example namespace
    """
    def __init__(self, public_endpoint: PublicEndpoint, ingress_src_file, ingress_name, ingress_host, ingress_pod_name,
                 namespace, ingress_event_text, ingress_error_event_text):
        self.public_endpoint = public_endpoint
        self.ingress_name = ingress_name
        self.ingress_pod_name = ingress_pod_name
        self.namespace = namespace
        self.ingress_host = ingress_host
        self.ingress_src_file = ingress_src_file
        self.ingress_event_text = ingress_event_text
        self.ingress_error_event_text = ingress_error_event_text


@pytest.fixture(scope="class")
def annotations_setup(request,
                      kube_apis,
                      ingress_controller_prerequisites,
                      ingress_controller_endpoint, ingress_controller, test_namespace) -> AnnotationsSetup:
    print("------------------------- Deploy Annotations-Example -----------------------------------")
    create_items_from_yaml(kube_apis,
                           f"{TEST_DATA}/annotations/{request.param}/annotations-ingress.yaml",
                           test_namespace)
    ingress_name = get_names_from_yaml(f"{TEST_DATA}/annotations/{request.param}/annotations-ingress.yaml")[0]
    ingress_host = get_first_ingress_host_from_yaml(f"{TEST_DATA}/annotations/{request.param}/annotations-ingress.yaml")
    common_app = create_example_app(kube_apis, "simple", test_namespace)
    wait_until_all_pods_are_ready(kube_apis.v1, test_namespace)
    ensure_connection_to_public_endpoint(ingress_controller_endpoint.public_ip,
                                         ingress_controller_endpoint.port,
                                         ingress_controller_endpoint.port_ssl)
    ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
    if request.param == 'mergeable':
        event_text = f"Configuration for {test_namespace}/{ingress_name}(Master) was added or updated"
        error_text = f"{event_text} but was not applied: Error reloading NGINX"
    else:
        event_text = f"Configuration for {test_namespace}/{ingress_name} was added or updated"
        error_text = f"{event_text}, but not applied: Error reloading NGINX"

    def fin():
        print("Clean up Annotations Example:")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    f"{DEPLOYMENTS}/common/nginx-config.yaml")
        delete_common_app(kube_apis.v1, kube_apis.apps_v1_api, common_app, test_namespace)
        delete_items_from_yaml(kube_apis,
                               f"{TEST_DATA}/annotations/{request.param}/annotations-ingress.yaml",
                               test_namespace)

    request.addfinalizer(fin)

    return AnnotationsSetup(ingress_controller_endpoint,
                            f"{TEST_DATA}/annotations/{request.param}/annotations-ingress.yaml",
                            ingress_name, ingress_host, ic_pod_name, test_namespace, event_text, error_text)


@pytest.fixture(scope="class")
def annotations_grpc_setup(request,
                           kube_apis,
                           ingress_controller_prerequisites,
                           ingress_controller_endpoint, ingress_controller, test_namespace) -> AnnotationsSetup:
    print("------------------------- Deploy gRPC Annotations-Example -----------------------------------")
    create_items_from_yaml(kube_apis,
                           f"{TEST_DATA}/annotations/grpc/annotations-ingress.yaml",
                           test_namespace)
    ingress_name = get_names_from_yaml(f"{TEST_DATA}/annotations/grpc/annotations-ingress.yaml")[0]
    ingress_host = get_first_ingress_host_from_yaml(f"{TEST_DATA}/annotations/grpc/annotations-ingress.yaml")
    replace_configmap_from_yaml(kube_apis.v1,
                                ingress_controller_prerequisites.config_map['metadata']['name'],
                                ingress_controller_prerequisites.namespace,
                                f"{TEST_DATA}/common/configmap-with-grpc.yaml")
    ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
    event_text = f"Configuration for {test_namespace}/{ingress_name} was added or updated"
    error_text = f"{event_text}, but not applied: Error reloading NGINX"

    def fin():
        print("Clean up gRPC Annotations Example:")
        delete_items_from_yaml(kube_apis,
                               f"{TEST_DATA}/annotations/grpc/annotations-ingress.yaml",
                               test_namespace)

    request.addfinalizer(fin)

    return AnnotationsSetup(ingress_controller_endpoint,
                            f"{TEST_DATA}/annotations/grpc/annotations-ingress.yaml",
                            ingress_name, ingress_host, ic_pod_name, test_namespace, event_text, error_text)


@pytest.mark.skip_for_nginx_plus
@pytest.mark.parametrize('annotations_setup', ["standard", "mergeable"], indirect=True)
class TestOssOnlyAnnotations:
    def test_nginx_config_defaults(self, kube_apis, annotations_setup, ingress_controller_prerequisites):
        print("Case 1: no ConfigMap keys, no annotations in Ingress")
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)

        assert "max_conns=0;" in result_conf

    @pytest.mark.parametrize('annotations, expected_strings, unexpected_strings', [
        ({"nginx.org/max-conns": "1024"}, ["max_conns=1024"], [])
    ])
    def test_when_annotation_in_ing_only(self, kube_apis, annotations_setup, ingress_controller_prerequisites,
                                         annotations, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_setup.namespace)
        initial_count = get_event_count(annotations_setup.ingress_event_text, initial_events)
        new_ing = generate_ingresses_with_annotation(annotations_setup.ingress_src_file,
                                                     annotations)
        for ing in new_ing:
            # in mergeable case this will update master ingress only
            if ing['metadata']['name'] == annotations_setup.ingress_name:
                replace_ingress(kube_apis.extensions_v1_beta1,
                                annotations_setup.ingress_name, annotations_setup.namespace, ing)
        wait_before_test(1)
        new_events = get_events(kube_apis.v1, annotations_setup.namespace)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)

        assert_event_count_increased(annotations_setup.ingress_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf


@pytest.mark.parametrize('annotations_setup', ["standard", "mergeable"], indirect=True)
class TestAnnotations:
    def test_nginx_config_defaults(self, kube_apis, annotations_setup, ingress_controller_prerequisites):
        print("Case 1: no ConfigMap keys, no annotations in Ingress")
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)

        assert "proxy_send_timeout 60s;" in result_conf

    @pytest.mark.parametrize('annotations, expected_strings, unexpected_strings', [
        ({"nginx.org/proxy-send-timeout": "10s"}, ["proxy_send_timeout 10s;"], ["proxy_send_timeout 60s;"]),
    ])
    def test_when_annotation_in_ing_only(self, kube_apis, annotations_setup, ingress_controller_prerequisites,
                                         annotations, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_setup.namespace)
        initial_count = get_event_count(annotations_setup.ingress_event_text, initial_events)
        print("Case 2: no ConfigMap keys, annotations in Ingress only")
        new_ing = generate_ingresses_with_annotation(annotations_setup.ingress_src_file,
                                                     annotations)
        for ing in new_ing:
            # in mergeable case this will update master ingress only
            if ing['metadata']['name'] == annotations_setup.ingress_name:
                replace_ingress(kube_apis.extensions_v1_beta1,
                                annotations_setup.ingress_name, annotations_setup.namespace, ing)
        wait_before_test(1)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        new_events = get_events(kube_apis.v1, annotations_setup.namespace)

        assert_event_count_increased(annotations_setup.ingress_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf

    @pytest.mark.parametrize('configmap_file, expected_strings, unexpected_strings', [
        (f"{TEST_DATA}/annotations/configmap-with-keys.yaml", ["proxy_send_timeout 33s;"], ["proxy_send_timeout 60s;"]),
    ])
    def test_when_annotation_in_configmap_only(self, kube_apis, annotations_setup, ingress_controller_prerequisites,
                                               configmap_file, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_setup.namespace)
        initial_count = get_event_count(annotations_setup.ingress_event_text, initial_events)
        print("Case 3: keys in ConfigMap, no annotations in Ingress")
        replace_ingresses_from_yaml(kube_apis.extensions_v1_beta1, annotations_setup.namespace,
                                    annotations_setup.ingress_src_file)
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    configmap_file)
        wait_before_test(1)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        new_events = get_events(kube_apis.v1, annotations_setup.namespace)

        assert_event_count_increased(annotations_setup.ingress_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf

    @pytest.mark.parametrize('annotations, configmap_file, expected_strings, unexpected_strings', [
        ({"nginx.org/proxy-send-timeout": "10s"},
         f"{TEST_DATA}/annotations/configmap-with-keys.yaml",
         ["proxy_send_timeout 10s;"], ["proxy_send_timeout 33s;"]),
    ])
    def test_ing_overrides_configmap(self, kube_apis, annotations_setup, ingress_controller_prerequisites,
                                     annotations, configmap_file, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_setup.namespace)
        initial_count = get_event_count(annotations_setup.ingress_event_text, initial_events)
        print("Case 4: keys in ConfigMap, annotations in Ingress")
        new_ing = generate_ingresses_with_annotation(annotations_setup.ingress_src_file,
                                                     annotations)
        for ing in new_ing:
            # in mergeable case this will update master ingress only
            if ing['metadata']['name'] == annotations_setup.ingress_name:
                replace_ingress(kube_apis.extensions_v1_beta1,
                                annotations_setup.ingress_name, annotations_setup.namespace, ing)
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    configmap_file)
        wait_before_test(1)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        new_events = get_events(kube_apis.v1, annotations_setup.namespace)

        assert_event_count_increased(annotations_setup.ingress_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf

    @pytest.mark.parametrize('annotations, expected_strings, unexpected_strings', [
        ({"nginx.org/proxy-send-timeout": "invalid"}, ["proxy_send_timeout invalid;"], ["proxy_send_timeout 60s;"]),
    ])
    def test_validation(self, kube_apis, annotations_setup, ingress_controller_prerequisites,
                        annotations, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_setup.namespace)
        print("Case 6: IC doesn't validate, only nginx validates")
        initial_count = get_event_count(annotations_setup.ingress_error_event_text, initial_events)
        new_ing = generate_ingresses_with_annotation(annotations_setup.ingress_src_file,
                                                     annotations)
        for ing in new_ing:
            # in mergeable case this will update master ingress only
            if ing['metadata']['name'] == annotations_setup.ingress_name:
                replace_ingress(kube_apis.extensions_v1_beta1,
                                annotations_setup.ingress_name, annotations_setup.namespace, ing)
        wait_before_test(1)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        new_events = get_events(kube_apis.v1, annotations_setup.namespace)

        assert_event_count_increased(annotations_setup.ingress_error_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf


@pytest.mark.parametrize('annotations_setup', ["mergeable"], indirect=True)
class TestMergeableFlows:
    @pytest.mark.parametrize('yaml_file, expected_strings, unexpected_strings', [
        (f"{TEST_DATA}/annotations/mergeable/minion-annotations-differ.yaml",
         ["proxy_send_timeout 25s;", "proxy_send_timeout 33s;"], ["proxy_send_timeout 10s;"]),
    ])
    def test_minion_overrides_master(self, kube_apis, annotations_setup, ingress_controller_prerequisites,
                                     yaml_file, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_setup.namespace)
        initial_count = get_event_count(annotations_setup.ingress_event_text, initial_events)
        print("Case 7: minion annotation overrides master")
        replace_ingresses_from_yaml(kube_apis.extensions_v1_beta1, annotations_setup.namespace, yaml_file)
        wait_before_test(1)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_setup.namespace,
                                                      annotations_setup.ingress_name,
                                                      annotations_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        new_events = get_events(kube_apis.v1, annotations_setup.namespace)

        assert_event_count_increased(annotations_setup.ingress_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf


class TestGrpcFlows:
    @pytest.mark.parametrize('annotations, expected_strings, unexpected_strings', [
        ({"nginx.org/proxy-send-timeout": "10s"}, ["grpc_send_timeout 10s;"], ["proxy_send_timeout 60s;"]),
    ])
    def test_grpc_flow(self, kube_apis, annotations_grpc_setup, ingress_controller_prerequisites,
                       annotations, expected_strings, unexpected_strings):
        initial_events = get_events(kube_apis.v1, annotations_grpc_setup.namespace)
        initial_count = get_event_count(annotations_grpc_setup.ingress_event_text, initial_events)
        print("Case 5: grpc annotations override http ones")
        new_ing = generate_ingresses_with_annotation(annotations_grpc_setup.ingress_src_file,
                                                     annotations)
        for ing in new_ing:
            if ing['metadata']['name'] == annotations_grpc_setup.ingress_name:
                replace_ingress(kube_apis.extensions_v1_beta1,
                                annotations_grpc_setup.ingress_name, annotations_grpc_setup.namespace, ing)
        wait_before_test(1)
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      annotations_grpc_setup.namespace,
                                                      annotations_grpc_setup.ingress_name,
                                                      annotations_grpc_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        new_events = get_events(kube_apis.v1, annotations_grpc_setup.namespace)

        assert_event_count_increased(annotations_grpc_setup.ingress_event_text, initial_count, new_events)
        for _ in expected_strings:
            assert _ in result_conf
        for _ in unexpected_strings:
            assert _ not in result_conf
