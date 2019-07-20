"""Describe methods to work with yaml files"""

import yaml


def get_first_ingress_host_from_yaml(file) -> str:
    """
    Parse yaml file and return first spec.rules[0].host appeared.

    :param file: an absolute path to file
    :return: str
    """
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            return dep['spec']['rules'][0]['host']


def get_external_host_from_service_yaml(file) -> str:
    """
    Parse yaml file and return first spec.externalName appeared.

    :param file: an absolute path to file
    :return: str
    """
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            return dep['spec']['externalName']


def get_names_from_yaml(file) -> []:
    """
    Parse yaml file and return all the found metadata.name.

    :param file: an absolute path to file
    :return: []
    """
    res = []
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            res.append(dep['metadata']['name'])
    return res


def get_paths_from_vs_yaml(file) -> []:
    """
    Parse yaml file and return all the found spec.routes.path.

    :param file: an absolute path to file
    :return: []
    """
    res = []
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            for route in dep['spec']['routes']:
                res.append(route['path'])
    return res


def get_first_vs_host_from_yaml(file) -> str:
    """
    Parse yaml file and return first spec.host appeared.

    :param file: an absolute path to file
    :return: str
    """
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            return dep['spec']['host']


def get_configmap_fields_from_yaml(file) -> {}:
    """
    Parse yaml file and return a dict of ConfigMap data fields.

    :param file: an absolute path to a file
    :return: {}
    """
    with open(file) as f:
        dep = yaml.load(f)
        return dep['data']


def get_route_namespace_from_vs_yaml(file) -> []:
    """
    Parse yaml file and return namespaces of all spec.routes.route.

    :param file: an absolute path to file
    :return: []
    """
    res = []
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            for route in dep['spec']['routes']:
                res.append(route['route'].split('/')[0])
    return res


def get_paths_from_vsr_yaml(file) -> []:
    """
    Parse yaml file and return all the found spec.subroutes.path.

    :param file: an absolute path to file
    :return: []
    """
    res = []
    with open(file) as f:
        docs = yaml.load_all(f)
        for dep in docs:
            for route in dep['spec']['subroutes']:
                res.append(route['path'])
    return res
