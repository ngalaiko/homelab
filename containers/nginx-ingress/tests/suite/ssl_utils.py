"""Describe methods to handle ssl connections."""

import socket
import ssl

import OpenSSL


def get_certificate(ip_address, host, port, timeout=10) -> str:
    """
    Get tls certificate.

    :param ip_address:
    :param host:
    :param port:
    :param timeout:
    :return: str
    """
    context = ssl.create_default_context()
    context.check_hostname = False
    context.verify_mode = ssl.CERT_NONE
    conn = socket.create_connection((ip_address, port))
    server_hostname = host if ssl.HAS_SNI else None
    sock = context.wrap_socket(conn, server_hostname=server_hostname)
    sock.settimeout(timeout)
    try:
        der_cert = sock.getpeercert(True)
    finally:
        sock.close()
    return ssl.DER_cert_to_PEM_cert(der_cert)


def get_server_certificate_subject(ip_address, host, port=443) -> dict:
    """
    Get tls certificate subject object.

    :param port: default is 443
    :param ip_address:
    :param host:
    :return: dict
    """
    certificate = get_certificate(ip_address, host, port)
    x509 = OpenSSL.crypto.load_certificate(OpenSSL.crypto.FILETYPE_PEM, certificate)
    return dict(x509.get_subject().get_components())
