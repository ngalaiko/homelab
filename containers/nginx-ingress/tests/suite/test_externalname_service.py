import requests
import pytest

from settings import TEST_DATA
from suite.fixtures import PublicEndpoint
from suite.resources_utils import create_ingress_from_yaml
from suite.resources_utils import replace_configmap_from_yaml, create_service_from_yaml
from suite.resources_utils import replace_configmap, delete_ingress, delete_service, get_ingress_nginx_template_conf
from suite.resources_utils import get_first_pod_name, ensure_connection_to_public_endpoint, wait_before_test
from suite.yaml_utils import get_first_ingress_host_from_yaml, get_external_host_from_service_yaml


class ExternalNameSetup:
    """Encapsulate ExternalName example details.

    Attributes:
        public_endpoint: PublicEndpoint
        ingress_name:
        ingress_pod_name:
        ingress_host:
        service: external-name example service name
        external_host: external-name example external host
        namespace: external-name example namespace
    """
    def __init__(self, public_endpoint: PublicEndpoint, ingress_name, ingress_host, ingress_pod_name, service, external_host, namespace):
        self.public_endpoint = public_endpoint
        self.ingress_name = ingress_name
        self.ingress_pod_name = ingress_pod_name
        self.namespace = namespace
        self.ingress_host = ingress_host
        self.service = service
        self.external_host = external_host


@pytest.fixture(scope="class")
def external_name_setup(request,
                        kube_apis,
                        ingress_controller_prerequisites,
                        ingress_controller_endpoint, ingress_controller, test_namespace) -> ExternalNameSetup:
    print("------------------------- Deploy External-Name-Example -----------------------------------")
    ingress_name = create_ingress_from_yaml(kube_apis.extensions_v1_beta1, test_namespace,
                                            f"{TEST_DATA}/externalname-services/externalname-ingress.yaml")
    ingress_host = get_first_ingress_host_from_yaml(f"{TEST_DATA}/externalname-services/externalname-ingress.yaml")
    external_host = get_external_host_from_service_yaml(f"{TEST_DATA}/externalname-services/externalname-svc.yaml")
    config_map_name = ingress_controller_prerequisites.config_map["metadata"]["name"]
    replace_configmap_from_yaml(kube_apis.v1, config_map_name,
                                ingress_controller_prerequisites.namespace,
                                f"{TEST_DATA}/externalname-services/nginx-config.yaml")
    svc_name = create_service_from_yaml(kube_apis.v1,
                                        test_namespace, f"{TEST_DATA}/externalname-services/externalname-svc.yaml")
    ensure_connection_to_public_endpoint(ingress_controller_endpoint.public_ip,
                                         ingress_controller_endpoint.port,
                                         ingress_controller_endpoint.port_ssl)
    ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)

    def fin():
        print("Clean up External-Name-Example:")
        replace_configmap(kube_apis.v1, config_map_name,
                          ingress_controller_prerequisites.namespace,
                          ingress_controller_prerequisites.config_map)
        delete_ingress(kube_apis.extensions_v1_beta1, ingress_name, test_namespace)
        delete_service(kube_apis.v1, svc_name, test_namespace)

    request.addfinalizer(fin)

    return ExternalNameSetup(ingress_controller_endpoint,
                             ingress_name, ingress_host, ic_pod_name, svc_name, external_host, test_namespace)


@pytest.mark.skip_for_nginx_oss
class TestExternalNameService:
    def test_resolver(self, external_name_setup):
        wait_before_test()
        req_url = f"http://{external_name_setup.public_endpoint.public_ip}:{external_name_setup.public_endpoint.port}/"
        resp = requests.get(req_url, headers={"host": external_name_setup.ingress_host}, verify=False)
        assert resp.status_code == 502

    def test_ic_template_config_upstream_zone(self, kube_apis, ingress_controller_prerequisites,
                                              ingress_controller, external_name_setup):
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      external_name_setup.namespace,
                                                      external_name_setup.ingress_name,
                                                      external_name_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        line = f"zone {external_name_setup.namespace}-{external_name_setup.ingress_name}-{external_name_setup.ingress_host}-{external_name_setup.service}-80 256k;"
        assert line in result_conf

    def test_ic_template_config_upstream_rule(self, kube_apis, ingress_controller_prerequisites,
                                              ingress_controller, external_name_setup):
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      external_name_setup.namespace,
                                                      external_name_setup.ingress_name,
                                                      external_name_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        assert "random two least_conn;" in result_conf

    def test_ic_template_config_upstream_server(self, kube_apis, ingress_controller_prerequisites,
                                                ingress_controller, ingress_controller_endpoint, external_name_setup):
        result_conf = get_ingress_nginx_template_conf(kube_apis.v1,
                                                      external_name_setup.namespace,
                                                      external_name_setup.ingress_name,
                                                      external_name_setup.ingress_pod_name,
                                                      ingress_controller_prerequisites.namespace)
        assert f"server {external_name_setup.external_host}:80 max_fails=1 fail_timeout=10s resolve;" in result_conf
