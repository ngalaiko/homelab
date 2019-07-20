"""Describe methods to utilize the kubernetes-client."""

import time
import yaml
import pytest
import requests

from kubernetes.client import CoreV1Api, ExtensionsV1beta1Api, RbacAuthorizationV1beta1Api, V1Service, AppsV1Api
from kubernetes.client.rest import ApiException
from kubernetes.stream import stream
from kubernetes import client
from settings import TEST_DATA, RECONFIGURATION_DELAY, DEPLOYMENTS
from suite.yaml_utils import get_names_from_yaml


class RBACAuthorization:
    """
    Encapsulate RBAC details.

    Attributes:
        role (str): cluster role name
        binding (str): cluster role binding name
    """
    def __init__(self, role: str, binding: str):
        self.role = role
        self.binding = binding


def configure_rbac(rbac_v1_beta1: RbacAuthorizationV1beta1Api) -> RBACAuthorization:
    """
    Create cluster and binding.

    :param rbac_v1_beta1: RbacAuthorizationV1beta1Api
    :return: RBACAuthorization
    """
    with open(f'{DEPLOYMENTS}/rbac/rbac.yaml') as f:
        docs = yaml.load_all(f)
        role_name = ""
        binding_name = ""
        for dep in docs:
            if dep["kind"] == "ClusterRole":
                print("Create cluster role")
                role_name = dep['metadata']['name']
                rbac_v1_beta1.create_cluster_role(dep)
                print(f"Created role '{role_name}'")
            elif dep["kind"] == "ClusterRoleBinding":
                print("Create binding")
                binding_name = dep['metadata']['name']
                rbac_v1_beta1.create_cluster_role_binding(dep)
                print(f"Created binding '{binding_name}'")
        return RBACAuthorization(role_name, binding_name)


def patch_rbac(rbac_v1_beta1: RbacAuthorizationV1beta1Api, yaml_manifest) -> RBACAuthorization:
    """
    Patch a clusterrole and a binding.

    :param rbac_v1_beta1: RbacAuthorizationV1beta1Api
    :param yaml_manifest: an absolute path to yaml manifest
    :return: RBACAuthorization
    """
    with open(yaml_manifest) as f:
        docs = yaml.load_all(f)
        role_name = ""
        binding_name = ""
        for dep in docs:
            if dep["kind"] == "ClusterRole":
                print("Patch the cluster role")
                role_name = dep['metadata']['name']
                rbac_v1_beta1.patch_cluster_role(role_name, dep)
                print(f"Patched the role '{role_name}'")
            elif dep["kind"] == "ClusterRoleBinding":
                print("Patch the binding")
                binding_name = dep['metadata']['name']
                rbac_v1_beta1.patch_cluster_role_binding(binding_name, dep)
                print(f"Patched the binding '{binding_name}'")
        return RBACAuthorization(role_name, binding_name)


def cleanup_rbac(rbac_v1_beta1: RbacAuthorizationV1beta1Api, rbac: RBACAuthorization) -> None:
    """
    Delete binding and cluster role.

    :param rbac_v1_beta1: RbacAuthorizationV1beta1Api
    :param rbac: RBACAuthorization
    :return:
    """
    delete_options = client.V1DeleteOptions()
    print("Delete binding and cluster role")
    rbac_v1_beta1.delete_cluster_role_binding(rbac.binding, delete_options)
    rbac_v1_beta1.delete_cluster_role(rbac.role, delete_options)


def create_deployment_from_yaml(apps_v1_api: AppsV1Api, namespace, yaml_manifest) -> str:
    """
    Create a deployment based on yaml file.

    :param apps_v1_api: AppsV1Api
    :param namespace: namespace name
    :param yaml_manifest: absolute path to file
    :return: str
    """
    print(f"Load {yaml_manifest}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
    return create_deployment(apps_v1_api, namespace, dep)


def create_deployment(apps_v1_api: AppsV1Api, namespace, body) -> str:
    """
    Create a deployment based on a dict.

    :param apps_v1_api: AppsV1Api
    :param namespace: namespace name
    :param body: dict
    :return: str
    """
    print("Create a deployment:")
    apps_v1_api.create_namespaced_deployment(namespace, body)
    print(f"Deployment created with name '{body['metadata']['name']}'")
    return body['metadata']['name']


def create_deployment_with_name(apps_v1_api: AppsV1Api, namespace, name) -> str:
    """
    Create a deployment with a specific name based on common yaml file.

    :param apps_v1_api: AppsV1Api
    :param namespace: namespace name
    :param name:
    :return: str
    """
    print(f"Create a Deployment with a specific name")
    with open(f"{TEST_DATA}/common/backend1.yaml") as f:
        dep = yaml.safe_load(f)
        dep['metadata']['name'] = name
        dep['spec']['selector']['matchLabels']['app'] = name
        dep['spec']['template']['metadata']['labels']['app'] = name
        dep['spec']['template']['spec']['containers'][0]['name'] = name
        return create_deployment(apps_v1_api, namespace, dep)


def create_daemon_set(apps_v1_api: AppsV1Api, namespace, body) -> str:
    """
    Create a daemon-set based on a dict.

    :param apps_v1_api: AppsV1Api
    :param namespace: namespace name
    :param body: dict
    :return: str
    """
    print("Create a daemon-set:")
    apps_v1_api.create_namespaced_daemon_set(namespace, body)
    print(f"Daemon-Set created with name '{body['metadata']['name']}'")
    return body['metadata']['name']


def wait_until_all_pods_are_ready(v1: CoreV1Api, namespace) -> None:
    """
    Wait for all the pods to be 'ContainersReady'.

    :param v1: CoreV1Api
    :param namespace: namespace of a pod
    :return:
    """
    print("Start waiting for all pods in a namespace to be ContainersReady")
    counter = 0
    while not are_all_pods_in_ready_state(v1, namespace) and counter < 20:
        print("There are pods that are not ContainersReady. Wait for 4 sec...")
        time.sleep(4)
        counter = counter + 1
    if counter >= 20:
        pytest.fail("After several seconds the pods aren't ContainersReady. Exiting...")
    print("All pods are ContainersReady")


def get_first_pod_name(v1: CoreV1Api, namespace) -> str:
    """
    Return 1st pod_name in a list of pods in a namespace.

    :param v1: CoreV1Api
    :param namespace:
    :return: str
    """
    resp = v1.list_namespaced_pod(namespace)
    return resp.items[0].metadata.name


def are_all_pods_in_ready_state(v1: CoreV1Api, namespace) -> bool:
    """
    Check if all the pods have ContainersReady condition.

    :param v1: CoreV1Api
    :param namespace: namespace
    :return: bool
    """
    pods = v1.list_namespaced_pod(namespace)
    if not pods.items:
        return False
    pod_ready_amount = 0
    for pod in pods.items:
        if pod.status.conditions is None:
            return False
        for condition in pod.status.conditions:
            # wait for 'Ready' state instead of 'ContainersReady' for backwards compatibility with k8s 1.10
            if condition.type == 'ContainersReady' and condition.status == 'True':
                pod_ready_amount = pod_ready_amount + 1
                break
    return pod_ready_amount == len(pods.items)


def create_service_from_yaml(v1: CoreV1Api, namespace, yaml_manifest) -> str:
    """
    Create a service based on yaml file.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :param yaml_manifest: absolute path to file
    :return: str
    """
    print(f"Load {yaml_manifest}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
    return create_service(v1, namespace, dep)


def create_service(v1: CoreV1Api, namespace, body) -> str:
    """
    Create a service based on a dict.

    :param v1: CoreV1Api
    :param namespace: namespace
    :param body: a dict
    :return: str
    """
    print("Create a Service:")
    resp = v1.create_namespaced_service(namespace, body)
    print(f"Service created with name '{body['metadata']['name']}'")
    return resp.metadata.name


def create_service_with_name(v1: CoreV1Api, namespace, name) -> str:
    """
    Create a service with a specific name based on a common yaml manifest.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :param name: name
    :return: str
    """
    print(f"Create a Service with a specific name:")
    with open(f"{TEST_DATA}/common/backend1-svc.yaml") as f:
        dep = yaml.safe_load(f)
        dep['metadata']['name'] = name
        dep['spec']['selector']['app'] = name.replace("-svc", "")
        return create_service(v1, namespace, dep)


def get_service_node_ports(v1: CoreV1Api, name, namespace) -> (str, str):
    """
    Get service allocated node_ports.

    :param v1: CoreV1Api
    :param name:
    :param namespace:
    :return: (plain_port, ssl_port)
    """
    resp = v1.read_namespaced_service(name, namespace)
    assert len(resp.spec.ports) == 2, "There are not enough ports assigned to a service"
    return resp.spec.ports[0].node_port, resp.spec.ports[1].node_port


def wait_for_public_ip(v1: CoreV1Api, namespace: str) -> str:
    """
    Wait for LoadBalancer to get the public ip.

    :param v1: CoreV1Api
    :param namespace: namespace
    :return: str
    """
    resp = v1.list_namespaced_service(namespace)
    counter = 0
    while str(resp.items[0].status.load_balancer.ingress) == "None" and counter < 20:
        time.sleep(5)
        resp = v1.list_namespaced_service(namespace)
        counter = counter + 1
    if counter == 20:
        pytest.fail("After 100 seconds the LB still doesn't have a Public IP. Exiting...")
    print(f"Public IP ='{resp.items[0].status.load_balancer.ingress[0].ip}'")
    return str(resp.items[0].status.load_balancer.ingress[0].ip)


def create_secret_from_yaml(v1: CoreV1Api, namespace, yaml_manifest) -> str:
    """
    Create a secret based on yaml file.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :param yaml_manifest: an absolute path to file
    :return: str
    """
    print(f"Load {yaml_manifest}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
    return create_secret(v1, namespace, dep)


def create_secret(v1: CoreV1Api, namespace, body) -> str:
    """
    Create a secret based on a dict.

    :param v1: CoreV1Api
    :param namespace: namespace
    :param body: a dict
    :return: str
    """
    print("Create a secret:")
    v1.create_namespaced_secret(namespace, body)
    print(f"Secret created: {body['metadata']['name']}")
    return body['metadata']['name']


def replace_secret(v1: CoreV1Api, name, namespace, yaml_manifest) -> str:
    """
    Replace a secret based on yaml file.

    :param v1: CoreV1Api
    :param name: secret name
    :param namespace: namespace name
    :param yaml_manifest: an absolute path to file
    :return: str
    """
    print(f"Replace a secret: '{name}'' in a namespace: '{namespace}'")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
        v1.replace_namespaced_secret(name, namespace, dep)
        print("Secret replaced")
    return name


def is_secret_present(v1: CoreV1Api, name, namespace) -> bool:
    """
    Check if a namespace has a secret.

    :param v1: CoreV1Api
    :param name:
    :param namespace:
    :return: bool
    """
    try:
        v1.read_namespaced_secret(name, namespace)
    except ApiException as ex:
        if ex.status == 404:
            print(f"No secret '{name}' found.")
            return False
    return True


def delete_secret(v1: CoreV1Api, name, namespace) -> None:
    """
    Delete a secret.

    :param v1: CoreV1Api
    :param name: secret name
    :param namespace: namespace name
    :return:
    """
    delete_options = client.V1DeleteOptions()
    delete_options.grace_period_seconds = 0
    delete_options.propagation_policy = 'Foreground'
    print(f"Delete a secret: {name}")
    v1.delete_namespaced_secret(name, namespace, delete_options)
    ensure_item_removal(v1.read_namespaced_secret, name, namespace)
    print(f"Secret was removed with name '{name}'")


def ensure_item_removal(get_item, *args, **kwargs) -> None:
    """
    Wait for item to be removed.

    :param get_item: a call to get an item
    :param args: *args
    :param kwargs: **kwargs
    :return:
    """
    try:
        counter = 0
        while counter < 30:
            time.sleep(1)
            get_item(*args, **kwargs)
            counter = counter + 1
        if counter >= 30:
            pytest.fail("Failed to remove the item after 30 seconds")
    except ApiException as ex:
        if ex.status == 404:
            print("Item was removed")


def create_ingress_from_yaml(extensions_v1_beta1: ExtensionsV1beta1Api, namespace, yaml_manifest) -> str:
    """
    Create an ingress based on yaml file.

    :param extensions_v1_beta1: ExtensionsV1beta1Api
    :param namespace: namespace name
    :param yaml_manifest: an absolute path to file
    :return: str
    """
    print(f"Load {yaml_manifest}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
        return create_ingress(extensions_v1_beta1, namespace, dep)


def create_ingress(extensions_v1_beta1: ExtensionsV1beta1Api, namespace, body) -> str:
    """
    Create an ingress based on a dict.

    :param extensions_v1_beta1: ExtensionsV1beta1Api
    :param namespace: namespace name
    :param body: a dict
    :return: str
    """
    print("Create an ingress:")
    extensions_v1_beta1.create_namespaced_ingress(namespace, body)
    print(f"Ingress created with name '{body['metadata']['name']}'")
    return body['metadata']['name']


def delete_ingress(extensions_v1_beta1: ExtensionsV1beta1Api, name, namespace) -> None:
    """
    Delete an ingress.

    :param extensions_v1_beta1: ExtensionsV1beta1Api
    :param namespace: namespace
    :param name:
    :return:
    """
    print(f"Delete an ingress: {name}")
    delete_options = client.V1DeleteOptions()
    extensions_v1_beta1.delete_namespaced_ingress(name, namespace, delete_options)
    ensure_item_removal(extensions_v1_beta1.read_namespaced_ingress, name, namespace)
    print(f"Ingress was removed with name '{name}'")


def create_namespace_from_yaml(v1: CoreV1Api, yaml_manifest) -> str:
    """
    Create a namespace based on yaml file.

    :param v1: CoreV1Api
    :param yaml_manifest: an absolute path to file
    :return: str
    """
    print(f"Load {yaml_manifest}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
        create_namespace(v1, dep)
        return dep['metadata']['name']


def create_namespace(v1: CoreV1Api, body) -> str:
    """
    Create an ingress based on a dict.

    :param v1: CoreV1Api
    :param body: a dict
    :return: str
    """
    print("Create a namespace:")
    v1.create_namespace(body)
    print(f"Namespace created with name '{body['metadata']['name']}'")
    return body['metadata']['name']


def create_namespace_with_name_from_yaml(v1: CoreV1Api, name, yaml_manifest) -> str:
    """
    Create a namespace with a specific name based on a yaml manifest.

    :param v1: CoreV1Api
    :param name: name
    :param yaml_manifest: an absolute path to file
    :return: str
    """
    print(f"Create a namespace with specific name:")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
        dep['metadata']['name'] = name
        v1.create_namespace(dep)
        print(f"Namespace created with name '{str(dep['metadata']['name'])}'")
        return dep['metadata']['name']


def create_service_account(v1: CoreV1Api, namespace, body) -> None:
    """
    Create a ServiceAccount based on a dict.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :param body: a dict
    :return:
    """
    print("Create a SA:")
    v1.create_namespaced_service_account(namespace, body)
    print(f"Service account created with name '{body['metadata']['name']}'")


def create_configmap_from_yaml(v1: CoreV1Api, namespace, yaml_manifest) -> str:
    """
    Create a config-map based on yaml file.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :param yaml_manifest: an absolute path to file
    :return: str
    """
    print(f"Load {yaml_manifest}")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
    return create_configmap(v1, namespace, dep)


def create_configmap(v1: CoreV1Api, namespace, body) -> str:
    """
    Create a config-map based on a dict.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :param body: a dict
    :return: str
    """
    print("Create a configMap:")
    v1.create_namespaced_config_map(namespace, body)
    print(f"Config map created with name '{body['metadata']['name']}'")
    return body["metadata"]["name"]


def replace_configmap_from_yaml(v1: CoreV1Api, name, namespace, yaml_manifest) -> None:
    """
    Replace a config-map based on a yaml file.

    :param v1: CoreV1Api
    :param name:
    :param namespace: namespace name
    :param yaml_manifest: an absolute path to file
    :return:
    """
    print(f"Replace a configMap: '{name}'")
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
        v1.replace_namespaced_config_map(name, namespace, dep)
        print("ConfigMap replaced")


def replace_configmap(v1: CoreV1Api, name, namespace, body) -> None:
    """
    Replace a config-map based on a dict.

    :param v1: CoreV1Api
    :param name:
    :param namespace:
    :param body: a dict
    :return:
    """
    print(f"Replace a configMap: '{name}'")
    v1.replace_namespaced_config_map(name, namespace, body)
    print("ConfigMap replaced")


def delete_namespace(v1: CoreV1Api, namespace) -> None:
    """
    Delete a namespace.

    :param v1: CoreV1Api
    :param namespace: namespace name
    :return:
    """
    print(f"Delete a namespace: {namespace}")
    delete_options = client.V1DeleteOptions()
    delete_options.grace_period_seconds = 0
    delete_options.propagation_policy = 'Foreground'
    v1.delete_namespace(namespace, delete_options)
    ensure_item_removal(v1.read_namespace, namespace)
    print(f"Namespace was removed with name '{namespace}'")


def get_file_contents(v1: CoreV1Api, file_path, pod_name, pod_namespace) -> str:
    """
    Execute 'cat file_path' command in a pod.

    :param v1: CoreV1Api
    :param pod_name: pod name
    :param pod_namespace: pod namespace
    :param file_path: an absolute path to a file in the pod
    :return: str
    """
    command = ["cat", file_path]
    resp = stream(
        v1.connect_get_namespaced_pod_exec,
        pod_name,
        pod_namespace,
        command=command,
        stderr=True, stdin=False, stdout=True, tty=False)
    result_conf = str(resp)
    print("\nFile contents:\n" + result_conf)
    return result_conf


def get_ingress_nginx_template_conf(v1: CoreV1Api, ingress_namespace, ingress_name, pod_name, pod_namespace) -> str:
    """
    Get contents of /etc/nginx/conf.d/{namespace}-{ingress_name}.conf in the pod.

    :param v1: CoreV1Api
    :param ingress_namespace:
    :param ingress_name:
    :param pod_name:
    :param pod_namespace:
    :return: str
    """
    file_path = f"/etc/nginx/conf.d/{ingress_namespace}-{ingress_name}.conf"
    return get_file_contents(v1, file_path, pod_name, pod_namespace)


class CommonApp:
    """
    Encapsulate Common Application details.

    Attributes:
        services ([]): list of services names
        deployments ([]): list of deployments names
    """
    def __init__(self, services, deployments):
        self.services = services
        self.deployments = deployments


def create_example_app(kube_apis, app_type, namespace) -> CommonApp:
    """
    Create a backend application.

    An application consists of 3 backend services.

    :param kube_apis: client apis
    :param app_type: type of the application (simple|split)
    :param namespace: namespace name
    :return: CommonApp
    """
    create_items_from_yaml(kube_apis, f"{TEST_DATA}/common/app/{app_type}/app.yaml", namespace)
    all_names = get_names_from_yaml(f"{TEST_DATA}/common/app/{app_type}/app.yaml")
    service_names = []
    dep_names = []
    for _ in all_names:
        if "svc" in _:
            service_names.append(_)
        else:
            dep_names.append(_)
    return CommonApp(service_names, dep_names)


def delete_common_app(v1: CoreV1Api,
                      apps_v1_api: AppsV1Api,
                      common_app: CommonApp, namespace) -> None:
    """
    Delete a common simple application.

    :param v1: CoreV1Api
    :param apps_v1_api: AppsV1Api
    :param common_app: CommonApp
    :param namespace: namespace name
    :return:
    """
    for deployment in common_app.deployments:
        delete_deployment(apps_v1_api, deployment, namespace)
    for svc in common_app.services:
        delete_service(v1, svc, namespace)


def delete_service(v1: CoreV1Api, name, namespace) -> None:
    """
    Delete a service.

    :param v1: CoreV1Api
    :param name:
    :param namespace:
    :return:
    """
    print(f"Delete a service: {name}")
    delete_options = client.V1DeleteOptions()
    delete_options.grace_period_seconds = 0
    delete_options.propagation_policy = 'Foreground'
    v1.delete_namespaced_service(name, namespace, delete_options)
    ensure_item_removal(v1.read_namespaced_service_status, name, namespace)
    print(f"Service was removed with name '{name}'")


def delete_deployment(apps_v1_api: AppsV1Api, name, namespace) -> None:
    """
    Delete a deployment.

    :param apps_v1_api: AppsV1Api
    :param name:
    :param namespace:
    :return:
    """
    print(f"Delete a deployment: {name}")
    delete_options = client.V1DeleteOptions()
    delete_options.grace_period_seconds = 0
    delete_options.propagation_policy = 'Foreground'
    apps_v1_api.delete_namespaced_deployment(name, namespace, delete_options)
    ensure_item_removal(apps_v1_api.read_namespaced_deployment_status, name, namespace)
    print(f"Deployment was removed with name '{name}'")


def delete_daemon_set(apps_v1_api: AppsV1Api, name, namespace) -> None:
    """
    Delete a daemon-set.

    :param apps_v1_api: AppsV1Api
    :param name:
    :param namespace:
    :return:
    """
    print(f"Delete a daemon-set: {name}")
    delete_options = client.V1DeleteOptions()
    delete_options.grace_period_seconds = 0
    delete_options.propagation_policy = 'Foreground'
    apps_v1_api.delete_namespaced_daemon_set(name, namespace, delete_options)
    ensure_item_removal(apps_v1_api.read_namespaced_daemon_set_status, name, namespace)
    print(f"Daemon-set was removed with name '{name}'")


def wait_before_test(delay=RECONFIGURATION_DELAY) -> None:
    """
    Wait for a time in seconds.

    :param delay: a delay in seconds
    :return:
    """
    time.sleep(delay)


def create_ingress_controller(v1: CoreV1Api, apps_v1_api: AppsV1Api, cli_arguments,
                              namespace, args=None) -> str:
    """
    Create an Ingress Controller according to the params.

    :param v1: CoreV1Api
    :param apps_v1_api: AppsV1Api
    :param cli_arguments: context name as in kubeconfig
    :param namespace: namespace name
    :param args: a list of any extra cli arguments to start IC with
    :return: str
    """
    print(f"Create an Ingress Controller as {cli_arguments['ic-type']}")
    yaml_manifest = f"{DEPLOYMENTS}/{cli_arguments['deployment-type']}/{cli_arguments['ic-type']}.yaml"
    with open(yaml_manifest) as f:
        dep = yaml.safe_load(f)
    dep['spec']['template']['spec']['containers'][0]['image'] = cli_arguments["image"]
    dep['spec']['template']['spec']['containers'][0]['imagePullPolicy'] = cli_arguments["image-pull-policy"]
    if args is not None:
        dep['spec']['template']['spec']['containers'][0]['args'].extend(args)
    if cli_arguments['deployment-type'] == 'deployment':
        name = create_deployment(apps_v1_api, namespace, dep)
    else:
        name = create_daemon_set(apps_v1_api, namespace, dep)
    wait_until_all_pods_are_ready(v1, namespace)
    print(f"Ingress Controller was created with name '{name}'")
    return name


def delete_ingress_controller(apps_v1_api: AppsV1Api, name, dep_type, namespace) -> None:
    """
    Delete IC according to its type.

    :param apps_v1_api: ExtensionsV1beta1Api
    :param name: name
    :param dep_type: IC deployment type 'deployment' or 'daemon-set'
    :param namespace: namespace name
    :return:
    """
    if dep_type == 'deployment':
        delete_deployment(apps_v1_api, name, namespace)
    elif dep_type == 'daemon-set':
        delete_daemon_set(apps_v1_api, name, namespace)


def create_ns_and_sa_from_yaml(v1: CoreV1Api, yaml_manifest) -> str:
    """
    Create a namespace and a service account in that namespace.

    :param v1:
    :param yaml_manifest: an absolute path to a file
    :return: str
    """
    print("Load yaml:")
    res = {}
    with open(yaml_manifest) as f:
        docs = yaml.load_all(f)
        for doc in docs:
            if doc["kind"] == "Namespace":
                res['namespace'] = create_namespace(v1, doc)
            elif doc["kind"] == "ServiceAccount":
                assert res['namespace'] is not None, "Ensure 'Namespace' is above 'SA' in the yaml manifest"
                create_service_account(v1, res['namespace'], doc)
    return res["namespace"]


def create_items_from_yaml(kube_apis, yaml_manifest, namespace) -> None:
    """
    Apply yaml manifest with multiple items.

    :param kube_apis: KubeApis
    :param yaml_manifest: an absolute path to a file
    :param namespace:
    :return:
    """
    print("Load yaml:")
    with open(yaml_manifest) as f:
        docs = yaml.load_all(f)
        for doc in docs:
            if doc["kind"] == "Secret":
                create_secret(kube_apis.v1, namespace, doc)
            elif doc["kind"] == "ConfigMap":
                create_configmap(kube_apis.v1, namespace, doc)
            elif doc["kind"] == "Ingress":
                create_ingress(kube_apis.extensions_v1_beta1, namespace, doc)
            elif doc["kind"] == "Service":
                create_service(kube_apis.v1, namespace, doc)
            elif doc["kind"] == "Deployment":
                create_deployment(kube_apis.apps_v1_api, namespace, doc)
            elif doc["kind"] == "DaemonSet":
                create_daemon_set(kube_apis.apps_v1_api, namespace, doc)


def delete_items_from_yaml(kube_apis, yaml_manifest, namespace) -> None:
    """
    Delete all the items found in the yaml file.

    :param kube_apis: KubeApis
    :param yaml_manifest: an absolute path to a file
    :param namespace: namespace
    :return:
    """
    print("Load yaml:")
    with open(yaml_manifest) as f:
        docs = yaml.load_all(f)
        for doc in docs:
            if doc["kind"] == "Namespace":
                delete_namespace(kube_apis.v1, doc['metadata']['name'])
            elif doc["kind"] == "Secret":
                delete_secret(kube_apis.v1, doc['metadata']['name'], namespace)
            elif doc["kind"] == "Ingress":
                delete_ingress(kube_apis.extensions_v1_beta1, doc['metadata']['name'], namespace)
            elif doc["kind"] == "Service":
                delete_service(kube_apis.v1, doc['metadata']['name'], namespace)
            elif doc["kind"] == "Deployment":
                delete_deployment(kube_apis.apps_v1_api, doc['metadata']['name'], namespace)
            elif doc["kind"] == "DaemonSet":
                delete_daemon_set(kube_apis.apps_v1_api, doc['metadata']['name'], namespace)


def ensure_connection(request_url) -> None:
    """
    Wait for connection.

    :param request_url: url to request
    :return:
    """
    for _ in range(4):
        try:
            resp = requests.get(request_url, verify=False)
            if resp.status_code == 404:
                return
        except Exception as ex:
            print(f"Warning: there was an exception {str(ex)}")
        time.sleep(3)
    pytest.fail("Connection failed after several attempts")


def ensure_connection_to_public_endpoint(ip_address, port, port_ssl) -> None:
    """
    Ensure the public endpoint doesn't refuse connections.

    :param ip_address:
    :param port:
    :param port_ssl:
    :return:
    """
    ensure_connection(f"http://{ip_address}:{port}/")
    ensure_connection(f"https://{ip_address}:{port_ssl}/")


def read_service(v1: CoreV1Api, name, namespace) -> V1Service:
    """
    Get details of a Service.

    :param v1: CoreV1Api
    :param name: service name
    :param namespace: namespace name
    :return: V1Service
    """
    print(f"Read a service named '{name}'")
    return v1.read_namespaced_service(name, namespace)


def replace_service(v1: CoreV1Api, name, namespace, body) -> str:
    """
    Patch a service based on a dict.

    :param v1: CoreV1Api
    :param name:
    :param namespace: namespace
    :param body: a dict
    :return: str
    """
    print(f"Replace a Service: {name}")
    resp = v1.replace_namespaced_service(name, namespace, body)
    print(f"Service updated with name '{name}'")
    return resp.metadata.name


def get_events(v1: CoreV1Api, namespace) -> []:
    """
    Get the list of events in a namespace.

    :param v1: CoreV1Api
    :param namespace:
    :return: []
    """
    print(f"Get the events in the namespace: {namespace}")
    res = v1.list_namespaced_event(namespace)
    return res.items


def ensure_response_from_backend(req_url, host) -> None:
    """
    Wait for 502 to disappear.

    :param req_url: url to request
    :param host:
    :return:
    """
    for _ in range(5):
        resp = requests.get(req_url, headers={"host": host}, verify=False)
        if resp.status_code != 502:
            return
        time.sleep(2)
    pytest.fail(f"Keep getting 502 from {req_url} after 10 seconds. Exiting...")
