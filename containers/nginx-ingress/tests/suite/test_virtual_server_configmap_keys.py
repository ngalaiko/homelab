import pytest

from settings import TEST_DATA, DEPLOYMENTS
from suite.resources_utils import wait_before_test, replace_configmap_from_yaml, get_events, get_first_pod_name
from suite.custom_resources_utils import get_vs_nginx_template_conf
from suite.yaml_utils import get_configmap_fields_from_yaml


def assert_valid_event_emitted(virtual_server_setup, new_list, previous_list):
    text_valid = f"Configuration for {virtual_server_setup.namespace}/{virtual_server_setup.vs_name} was updated"
    text_invalid = "was updated but was not applied"
    new_event = new_list[len(new_list) - 1]
    assert len(new_list) - len(previous_list) == 1
    assert text_valid in new_event.message and text_invalid not in new_event.message


def assert_invalid_event_emitted(virtual_server_setup, new_list, previous_list):
    text_invalid = f"Configuration for {virtual_server_setup.namespace}/{virtual_server_setup.vs_name} was updated but was not applied"
    new_event = new_list[len(new_list) - 1]
    assert len(new_list) - len(previous_list) == 1
    assert text_invalid in new_event.message


def assert_valid_event_count_increased(virtual_server_setup, new_list, previous_list):
    text_valid = f"Configuration for {virtual_server_setup.namespace}/{virtual_server_setup.vs_name} was updated"
    text_invalid = "was updated but was not applied"
    for i in range(len(previous_list)-1, 0, -1):
        if text_valid in previous_list[i].message and text_invalid not in previous_list[i].message:
            assert new_list[i].count - previous_list[i].count == 1, "We expect the counter to increase"


def assert_keys_without_validation(config, expected_values):
    assert f"proxy_connect_timeout {expected_values['proxy-connect-timeout']};" in config
    assert f"proxy_read_timeout {expected_values['proxy-read-timeout']};" in config
    assert f"client_max_body_size {expected_values['client-max-body-size']};" in config
    assert f"proxy_buffers {expected_values['proxy-buffers']};" in config
    assert f"proxy_buffer_size {expected_values['proxy-buffer-size']};" in config
    assert f"proxy_max_temp_file_size {expected_values['proxy-max-temp-file-size']};" in config
    assert f"set_real_ip_from {expected_values['set-real-ip-from']};" in config
    assert f"real_ip_header {expected_values['real-ip-header']};" in config
    assert f"{expected_values['location-snippets']}" in config
    assert f"{expected_values['server-snippets']}" in config
    assert f"fail_timeout={expected_values['fail-timeout']};" in config
    assert f"proxy_send_timeout {expected_values['proxy-send-timeout']};" in config


def assert_keys_with_validation(config, expected_values):
    # based on f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys.yaml"
    assert "proxy_buffering off;" in config
    assert "real_ip_recursive on;" in config
    assert f"max_fails={expected_values['max-fails']}" in config
    assert f"keepalive {expected_values['keepalive']};" in config
    assert "listen 80 proxy_protocol;" in config
    assert "if ($http_x_forwarded_proto = 'http') {" in config


def assert_specific_keys_for_nginx_plus(config, expected_values):
    # based on f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys.yaml"
    assert f"server_tokens \"{expected_values['server-tokens']}\";" in config
    assert "random two least_conn;" not in config \
           and expected_values['lb-method'] in config


def assert_specific_keys_for_nginx_oss(config, expected_values):
    # based on f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys-oss.yaml"
    assert "server_tokens \"off\"" in config
    assert "random two least_conn;" not in config \
           and expected_values['lb-method'] in config


def assert_defaults_of_keys_with_validation(config, unexpected_values):
    assert "proxy_buffering on;" in config
    assert "real_ip_recursive" not in config
    assert "max_fails=1" in config
    assert "keepalive" not in config
    assert "listen 80;" in config
    assert "if ($http_x_forwarded_proto = 'http') {" not in config
    assert "server_tokens \"on\"" in config
    assert "random two least_conn;" in config and unexpected_values['lb-method'] not in config
    assert f"proxy_send_timeout 60s;" in config


def assert_ssl_keys(config):
    # based on f"{TEST_DATA}/virtual-server-configmap-keys/configmap-ssl-keys.yaml"
    assert "if ($schema = 'http') {" not in config
    assert "listen 443 ssl http2 proxy_protocol;" in config


def assert_defaults_of_ssl_keys(config):
    assert "if ($schema = 'http') {" not in config
    assert "listen 443 ssl;" in config
    assert "http2" not in config


@pytest.fixture(scope="function")
def clean_up(request, kube_apis, ingress_controller_prerequisites, test_namespace) -> None:
    """
    Return ConfigMap to the initial state after the test.

    :param request: internal pytest fixture
    :param kube_apis: client apis
    :param ingress_controller_prerequisites:
    :param test_namespace: str
    :return:
    """

    def fin():
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    f"{DEPLOYMENTS}/common/nginx-config.yaml")

    request.addfinalizer(fin)


@pytest.mark.parametrize('crd_ingress_controller, virtual_server_setup',
                         [({"type": "complete", "extra_args": [f"-enable-custom-resources"]},
                           {"example": "virtual-server-configmap-keys", "app_type": "simple"})],
                         indirect=True)
class TestVirtualServerConfigMapNoTls:
    def test_keys(self, cli_arguments, kube_apis, ingress_controller_prerequisites,
                  crd_ingress_controller, virtual_server_setup, clean_up):
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        initial_list = get_events(kube_apis.v1, virtual_server_setup.namespace)

        print("Step 1: update ConfigMap with valid keys without validation rules")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    f"{TEST_DATA}/virtual-server-configmap-keys/configmap-no-validation-keys.yaml")
        expected_values = get_configmap_fields_from_yaml(
            f"{TEST_DATA}/virtual-server-configmap-keys/configmap-no-validation-keys.yaml")
        wait_before_test(1)
        step_1_events = get_events(kube_apis.v1, virtual_server_setup.namespace)
        step_1_config = get_vs_nginx_template_conf(kube_apis.v1,
                                                   virtual_server_setup.namespace,
                                                   virtual_server_setup.vs_name,
                                                   ic_pod_name,
                                                   ingress_controller_prerequisites.namespace)
        assert_valid_event_emitted(virtual_server_setup, step_1_events, initial_list)
        assert_keys_without_validation(step_1_config, expected_values)

        print("Step 2: update ConfigMap with invalid keys without validation rules")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    f"{TEST_DATA}/virtual-server-configmap-keys/configmap-no-validation-keys-invalid.yaml")
        expected_values = get_configmap_fields_from_yaml(
            f"{TEST_DATA}/virtual-server-configmap-keys/configmap-no-validation-keys-invalid.yaml")
        wait_before_test(1)
        step_2_events = get_events(kube_apis.v1, virtual_server_setup.namespace)
        step_2_config = get_vs_nginx_template_conf(kube_apis.v1,
                                                   virtual_server_setup.namespace,
                                                   virtual_server_setup.vs_name,
                                                   ic_pod_name,
                                                   ingress_controller_prerequisites.namespace)
        assert_invalid_event_emitted(virtual_server_setup, step_2_events, step_1_events)
        assert_keys_without_validation(step_2_config, expected_values)

        # to cover the OSS case, this will be changed in the future
        if cli_arguments['ic-type'] == "nginx-ingress":
            data_file = f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys-oss.yaml"
            data_file_invalid = f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys-invalid-oss.yaml"
        else:
            data_file = f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys.yaml"
            data_file_invalid = f"{TEST_DATA}/virtual-server-configmap-keys/configmap-validation-keys-invalid.yaml"

        print("Step 3: update ConfigMap with valid keys with validation rules")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    data_file)
        expected_values = get_configmap_fields_from_yaml(data_file)
        wait_before_test(1)
        step_3_config = get_vs_nginx_template_conf(kube_apis.v1,
                                                   virtual_server_setup.namespace,
                                                   virtual_server_setup.vs_name,
                                                   ic_pod_name,
                                                   ingress_controller_prerequisites.namespace)
        step_3_events = get_events(kube_apis.v1, virtual_server_setup.namespace)
        assert_valid_event_count_increased(virtual_server_setup, step_3_events, step_2_events)
        assert_keys_with_validation(step_3_config, expected_values)
        # to cover the OSS case, this will be changed in the future
        if cli_arguments['ic-type'] == "nginx-ingress":
            assert_specific_keys_for_nginx_oss(step_3_config, expected_values)
        else:
            assert_specific_keys_for_nginx_plus(step_3_config, expected_values)

        print("Step 4: update ConfigMap with invalid keys")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    data_file_invalid)
        expected_values = get_configmap_fields_from_yaml(data_file_invalid)
        wait_before_test(1)
        step_4_config = get_vs_nginx_template_conf(kube_apis.v1,
                                                   virtual_server_setup.namespace,
                                                   virtual_server_setup.vs_name,
                                                   ic_pod_name,
                                                   ingress_controller_prerequisites.namespace)
        step_4_events = get_events(kube_apis.v1, virtual_server_setup.namespace)
        assert_valid_event_count_increased(virtual_server_setup, step_4_events, step_3_events)
        assert_defaults_of_keys_with_validation(step_4_config, expected_values)


@pytest.mark.parametrize('crd_ingress_controller, virtual_server_setup',
                         [({"type": "complete", "extra_args": [f"-enable-custom-resources"]},
                           {"example": "virtual-server-tls", "app_type": "simple"})],
                         indirect=True)
class TestVirtualServerConfigMapWithTls:
    def test_ssl_keys(self, kube_apis, ingress_controller_prerequisites, crd_ingress_controller,
                      virtual_server_setup, clean_up):
        ic_pod_name = get_first_pod_name(kube_apis.v1, ingress_controller_prerequisites.namespace)
        initial_list = get_events(kube_apis.v1, virtual_server_setup.namespace)

        print("Step 1: update ConfigMap with valid ssl keys")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    f"{TEST_DATA}/virtual-server-configmap-keys/configmap-ssl-keys.yaml")
        wait_before_test(1)
        step_1_events = get_events(kube_apis.v1, virtual_server_setup.namespace)
        step_1_config = get_vs_nginx_template_conf(kube_apis.v1,
                                                   virtual_server_setup.namespace,
                                                   virtual_server_setup.vs_name,
                                                   ic_pod_name,
                                                   ingress_controller_prerequisites.namespace)
        assert_valid_event_emitted(virtual_server_setup, step_1_events, initial_list)
        assert_ssl_keys(step_1_config)

        print("Step 2: update ConfigMap with invalid ssl keys")
        replace_configmap_from_yaml(kube_apis.v1,
                                    ingress_controller_prerequisites.config_map['metadata']['name'],
                                    ingress_controller_prerequisites.namespace,
                                    f"{TEST_DATA}/virtual-server-configmap-keys/configmap-ssl-keys-invalid.yaml")
        wait_before_test(1)
        step_2_events = get_events(kube_apis.v1, virtual_server_setup.namespace)
        step_2_config = get_vs_nginx_template_conf(kube_apis.v1,
                                                   virtual_server_setup.namespace,
                                                   virtual_server_setup.vs_name,
                                                   ic_pod_name,
                                                   ingress_controller_prerequisites.namespace)
        assert_valid_event_count_increased(virtual_server_setup, step_2_events, step_1_events)
        assert_defaults_of_ssl_keys(step_2_config)
