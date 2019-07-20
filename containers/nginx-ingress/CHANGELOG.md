# Changelog

### 1.5.1

CHANGES:
* Update NGINX version to 1.17.1.

HELM CHART:
* The version of the Helm chart is now 0.3.1.
* [593](https://github.com/nginxinc/kubernetes-ingress/pull/593): Fix the selector in the Ingress Controller service when the `controller.name` parameter is set. This introduces a change, see the HELM UPGRADE section.

UPGRADE:
* For NGINX, use the 1.5.1 image from our DockerHub: `nginx/nginx-ingress:1.5.1` or `nginx/nginx-ingress:1.5.1-alpine`
* For NGINX Plus, please build your own image using the 1.5.1 source code.
* For Helm, use version 0.3.1 of the chart.

HELM UPGRADE:

In the changelog of Release 1.5.0, we advised not to upgrade the helm chart from `0.2.1` to `0.3.0` unless the mentioned in the changelog problems were acceptable. This release we provide mitigation instructions on how to upgrade from `0.2.1` to `0.3.1` without disruptions. 

When you upgrade from `0.2.1` to `0.3.1`, make sure to configure the following parameters:
* `controller.name` is set to `nginx-ingress` or the previously used value in case you customized it. This ensures the Deployment/Daemonset will not be recreated.
* `controller.service.name` is set to `nginx-ingress`. This ensures the service will not be recreated.
* `controller.config.name` is set to `nginx-config`. This ensures the ConfigMap will not be recreated.

Upgrading from `0.3.0` to `0.3.1`: Upgrading is not affected unless you customized `controller.name`. In that case, because of the fix [593](https://github.com/nginxinc/kubernetes-ingress/pull/593), the upgraded service will have a new selector, and the upgraded pod spec will have a new label. As a result, during an upgrade, the old pods will be immediately excluded from the service. Also, for the Deployment, the old pods will not terminate but continue to run. To terminate the old pods, manually remove the corresponding ReplicaSet.

### 1.5.0

FEATURES:
* [560](https://github.com/nginxinc/kubernetes-ingress/pull/560): Add new configuration resources -- VirtualServer and VirtualServerRoute.
* [554](https://github.com/nginxinc/kubernetes-ingress/pull/554): Add new Prometheus metrics related to the Ingress Controller's operation (as opposed to NGINX/NGINX Plus metrics).
* [496](https://github.com/nginxinc/kubernetes-ingress/pull/496): Support a wildcard TLS certificate for TLS-enabled Ingress resources.
* [485](https://github.com/nginxinc/kubernetes-ingress/pull/485): Support ExternalName services in Ingress backends.

IMPROVEMENTS:
* Add new ConfigMap keys: `keepalive-timeout`, `keepalive-requests`, `access-log-off`, `variables-hash-bucket-size`, `variables-hash-max-size`. Added in [565](https://github.com/nginxinc/kubernetes-ingress/pull/565), [511](https://github.com/nginxinc/kubernetes-ingress/pull/511).
* [504](https://github.com/nginxinc/kubernetes-ingress/pull/504): Run the Prometheus exporter inside the Ingress Controller process instead of a sidecar container.

BUGFIXES:
* [520](https://github.com/nginxinc/kubernetes-ingress/pull/520): Fix the type of the Prometheus port annotation in manifests.
* [481](https://github.com/nginxinc/kubernetes-ingress/pull/481): Fix the HSTS support.
* [439](https://github.com/nginxinc/kubernetes-ingress/pull/439): Fix the validation of the `lb-method` ConfigMap key and `nginx.org/lb-method` annotation.

HELM CHART:
* The version of the helm chart is now 0.3.0.
* The helm chart is now available in our helm chart repo `helm.nginx.com/stable`. 
* Add new parameters to the Chart: `controller.service.httpPort.targetPort`, `controller.service.httpsPort.targetPort`, `controller.service.name`, `controller.pod.annotations`, `controller.config.name`, `controller.reportIngressStatus.leaderElectionLockName`, `controller.service.httpPort`, `controller.service.httpsPort`, `controller.service.loadBalancerIP`, `controller.service.loadBalancerSourceRanges`, `controller.tolerations`, `controller.affinity`. Added in [562](https://github.com/nginxinc/kubernetes-ingress/pull/562), [561](https://github.com/nginxinc/kubernetes-ingress/pull/561), [553](https://github.com/nginxinc/kubernetes-ingress/pull/553), [534](https://github.com/nginxinc/kubernetes-ingress/pull/534) thanks to [Paulo Ribeiro](https://github.com/paigr), [479](https://github.com/nginxinc/kubernetes-ingress/pull/479) thanks to [Alejandro Llanes](https://github.com/sombralibre),  [468](https://github.com/nginxinc/kubernetes-ingress/pull/468), [456](https://github.com/nginxinc/kubernetes-ingress/pull/456).
* [546](https://github.com/nginxinc/kubernetes-ingress/pull/546): Support deploying multiple Ingress Controllers in a cluster. **Note**: The generated resources have new names that are unique for each Ingress Controller. As a consequence, the name change affects the upgrade. See the HELM UPGRADE section for more information. 
* [542](https://github.com/nginxinc/kubernetes-ingress/pull/542): Reduce the required privileges in the RBAC manifests.

CHANGES:
* Update NGINX version to 1.15.12.
* Prometheus metrics for NGINX/NGINX Plus have new namespace `nginx_ingress`. Examples: `nginx_http_requests_total` -> `nginx_ingress_http_requests_total`, `nginxplus_http_requests_total` -> `nginx_ingress_nginxplus_http_requests_total`.

UPGRADE:
* For NGINX, use the 1.5.0 image from our DockerHub: `nginx/nginx-ingress:1.5.0` or `nginx/nginx-ingress:1.5.0-alpine`
* For NGINX Plus, please build your own image using the 1.5.0 source code.
* For Helm, use version 0.3.0 of the chart.

HELM UPGRADE:

The new version of the helm chart uses different names for the generated resources. This makes it possible to deploy multiple Ingress Controllers in a cluster. However, as a side effect, during the upgrade from the previous version, helm will recreate the resources, instead of updating the existing ones. This, in turn, might cause problems for the following resources:
* Service: If the service was created with the type LoadBalancer, the public IP of the new service might change. Additionally, helm updates the selector of the service, so that the old pods will be immediately excluded from the service. 
* Deployment/DaemonSet: Because the resource is recreated, the old pods will be removed and the new ones will be launched, instead of the default Deployment/Daemonset upgrade strategy. 
* ConfigMap: After the helm removes the resource, the old Ingress Controller pods will be immediately reconfigured to use the default values of the ConfigMap keys. During a small window between the reconfiguration and the shutdown of the old pods, NGINX will use the configuration with the default values.

We advise not to upgrade to the new version of the helm chart unless the mentioned problems are acceptable for your case. We will provide special upgrade instructions for helm that mitigate the problems for the next minor release of the Ingress Controller (1.5.1).

### 1.4.6

CHANGES:
* Update NGINX version to 1.15.11.
* Update NGINX Plus version to R18.

HELM CHART:
* The version of the Helm chart is now 0.2.1.

UPGRADE:
* For NGINX, use the 1.4.6 image from our DockerHub: `nginx/nginx-ingress:1.4.6` or `nginx/nginx-ingress:1.4.6-alpine`
* For NGINX Plus, please build your own image using the 1.4.6 source code.
* For Helm, use version 0.2.1 of the chart.

### 1.4.5

CHANGES:
* Update NGINX version to 1.15.10.

UPGRADE:
* For NGINX, use the 1.4.5 image from our DockerHub: `nginx/nginx-ingress:1.4.5` or `nginx/nginx-ingress:1.4.5-alpine`
* For NGINX Plus, please build your own image using the 1.4.5 source code.

### 1.4.4

CHANGES:
* Update NGINX version to 1.15.9.

UPGRADE:
* For NGINX, use the 1.4.4 image from our DockerHub: `nginx/nginx-ingress:1.4.4` or `nginx/nginx-ingress:1.4.4-alpine`
* For NGINX Plus, please build your own image using the 1.4.4 source code.

### 1.4.3

CHANGES:
* Update NGINX version to 1.15.8.

UPGRADE:
* For NGINX, use the 1.4.3 image from our DockerHub: `nginx/nginx-ingress:1.4.3` or `nginx/nginx-ingress:1.4.3-alpine`
* For NGINX Plus, please build your own image using the 1.4.3 source code.

### 1.4.2

CHANGES:
* Update NGINX Plus version to R17.

 UPGRADE:
* For NGINX, use the 1.4.2 image from our DockerHub: `nginx/nginx-ingress:1.4.2` or `nginx/nginx-ingress:1.4.2-alpine`
* For NGINX Plus, please build your own image using the 1.4.2 source code.

### 1.4.1

CHANGES:
* Update NGINX version to 1.15.7.

UPGRADE:
* For NGINX, use the 1.4.1 image from our DockerHub: `nginx/nginx-ingress:1.4.1` or `nginx/nginx-ingress:1.4.1-alpine`
* For NGINX Plus, please build your own image using the 1.4.1 source code.

### 1.4.0

FEATURES:
* [401](https://github.com/nginxinc/kubernetes-ingress/pull/401): Add the `-nginx-debug` flag for enabling debugging of NGINX using the `nginx-debug` binary.
* [387](https://github.com/nginxinc/kubernetes-ingress/pull/387): Add the `-nginx-status-allow-cidrs` command-line argument for white listing IPv4 IP/CIDR blocks to allow access to NGINX stub_status or the NGINX Plus API. Thanks to [Jasmine Hegman](https://github.com/r4j4h).
* [376](https://github.com/nginxinc/kubernetes-ingress/pull/376): Support the [random](http://nginx.org/en/docs/http/ngx_http_upstream_module.html#random) load balancing method.
* [375](https://github.com/nginxinc/kubernetes-ingress/pull/375): Support custom annotations.
* [346](https://github.com/nginxinc/kubernetes-ingress/pull/346): Support the Prometheus exporter for NGINX (the stub_status metrics).
* [344](https://github.com/nginxinc/kubernetes-ingress/pull/344): Expose NGINX Plus API/NGINX stub_status on a custom port via the `-nginx-status-port` command-line argument. See also the CHANGES section.
* [342](https://github.com/nginxinc/kubernetes-ingress/pull/342): Add the `error-log-level` configmap key. Thanks to [boran seref](https://github.com/boranx).
* [320](https://github.com/nginxinc/kubernetes-ingress/pull/340): Support TCP/UDP load balancing via the `stream-snippets` configmap key.

IMPROVEMENTS:
* [434](https://github.com/nginxinc/kubernetes-ingress/pull/434): Improve consistency of templates.
* [432](https://github.com/nginxinc/kubernetes-ingress/pull/432): Fix cli-docs and Improve main test.
* [419](https://github.com/nginxinc/kubernetes-ingress/pull/419): Refactor config writing. Thanks to [feifeiiiiiiiiii](https://github.com/feifeiiiiiiiiiii).
* [403](https://github.com/nginxinc/kubernetes-ingress/pull/403): Improve NGINX start.
* [400](https://github.com/nginxinc/kubernetes-ingress/pull/400): Fix error message in internal/controller/controller.go. Thanks to [Alex O Regan](https://github.com/aaaaaaaalex).
* [399](https://github.com/nginxinc/kubernetes-ingress/pull/399): Improve secret handling. See also the CHANGES section.
* [391](https://github.com/nginxinc/kubernetes-ingress/pull/391): Update default lb-method to be random two least_conn. See also the CHANGES section.
* [389](https://github.com/nginxinc/kubernetes-ingress/pull/389): Improve parsing nginx.org/rewrites annotation.
* [380](https://github.com/nginxinc/kubernetes-ingress/pull/380): Verify reloads & cache secrets.
* [362](https://github.com/nginxinc/kubernetes-ingress/pull/362): Reduce reloads.
* [357](https://github.com/nginxinc/kubernetes-ingress/pull/357): Improve Project Layout and Refactor Controller Package. See also the CHANGES section.
* [351](https://github.com/nginxinc/kubernetes-ingress/pull/351): Make socket address obvious.

BUGFIXES:
* [429](https://github.com/nginxinc/kubernetes-ingress/pull/429): Fix panic with health checks.
* [386](https://github.com/nginxinc/kubernetes-ingress/pull/386): Fix Configmap/Mergeable Ingress Add/Update event logging.
* [379](https://github.com/nginxinc/kubernetes-ingress/pull/379): Fix configmap update.
* [365](https://github.com/nginxinc/kubernetes-ingress/pull/365): Don't enqueue ingress for some service changes.
* [348](https://github.com/nginxinc/kubernetes-ingress/pull/348): Fix Configurator error check.

HELM CHART:
* [430](https://github.com/nginxinc/kubernetes-ingress/pull/430): Add the `controller.serviceAccount.imagePullSecrets` parameter to the helm chart. See also the CHANGES section.
* [420](https://github.com/nginxinc/kubernetes-ingress/pull/420): Simplify values files for Helm Chart.
* [398](https://github.com/nginxinc/kubernetes-ingress/pull/398): Add the `controller.nginxStatus.allowCidrs` and `controller.service.externalIPs` parameters to helm chart.
* [393](https://github.com/nginxinc/kubernetes-ingress/pull/393): Refactor Helm Chart templates.
* [390](https://github.com/nginxinc/kubernetes-ingress/pull/390): Add the `controller.service.loadBalancerIP` parameter to the helm chat.
* [377](https://github.com/nginxinc/kubernetes-ingress/pull/377): Add the `controller.nginxStatus` parameters to the helm chart.
* [335](https://github.com/nginxinc/kubernetes-ingress/pull/335): Add the `controller.reportIngressStatus` parameters to the helm chart.
* The version of the Helm chart is now 0.2.0.

CHANGES:
* Update NGINX version to 1.15.6.   
* Update NGINX Plus version to R16p1.
* Update NGINX Prometheus Exporter to 0.2.0.
* [430](https://github.com/nginxinc/kubernetes-ingress/pull/430): Add the `controller.serviceAccount.imagePullSecrets` parameter to the helm chart. **Note**: the `controller.serviceAccountName` parameter has been changed to `controller.serviceAccount.name`.
* [399](https://github.com/nginxinc/kubernetes-ingress/pull/399): Improve secret handling. **Note**: the PR changed how the Ingress Controller processes Ingress resources with TLS termination enabled but without any referenced (or with invalid) secrets and Ingress resources with JWT validation enabled but without any referenced (or with invalid) JWK. Please read [here](https://github.com/nginxinc/kubernetes-ingress/pull/399) for more details.
* [357](https://github.com/nginxinc/kubernetes-ingress/pull/357): Improve Project Layout and Refactor Controller Package. **Note**: the PR significantly changed the layout of the project to follow best practices.
* [347](https://github.com/nginxinc/kubernetes-ingress/pull/347): Use edge version in manifests and Helm chart. **Note**: the manifests and the helm chart in the master branch now reference the edge version of the Ingress Controller instead of the latest stable version used previously.
* [391](https://github.com/nginxinc/kubernetes-ingress/pull/391): Update default lb-method to be random two least_conn. **Note**: the default load balancing method is now the power of two choices as it better suits the Ingress Controller use case. Please read the [blog post](https://www.nginx.com/blog/nginx-power-of-two-choices-load-balancing-algorithm/) about the method for more details.
* [344](https://github.com/nginxinc/kubernetes-ingress/pull/344): Expose NGINX Plus API/NGINX stub_status on a custom port via the `-nginx-status-port` command-line argument. **Note**: For NGINX the stub_status is now exposed on port 8080 at the /stub_status URL by default. Previously, the stub_status was not exposed on any port. The stub_status can be disabled via the `-nginx-status` flag.

DOC AND EXAMPLES FIXES/IMPROVEMENTS: [435](https://github.com/nginxinc/kubernetes-ingress/pull/435), [433](https://github.com/nginxinc/kubernetes-ingress/pull/433), [432](https://github.com/nginxinc/kubernetes-ingress/pull/432), [418](https://github.com/nginxinc/kubernetes-ingress/pull/418) (Thanks to [Hal Deadman](https://github.com/hdeadman)), [406](https://github.com/nginxinc/kubernetes-ingress/pull/406),  [381](https://github.com/nginxinc/kubernetes-ingress/pull/381), [349](https://github.com/nginxinc/kubernetes-ingress/pull/349) (Thanks to [Artur Geraschenko](https://github.com/arturgspb)), [343](https://github.com/nginxinc/kubernetes-ingress/pull/343)

UPGRADE:
* For NGINX, use the 1.4.0 image from our DockerHub: `nginx/nginx-ingress:1.4.0` or `nginx/nginx-ingress:1.4.0-alpine`
* For NGINX Plus, please build your own image using the 1.4.0 source code.

### 1.3.2

CHANGES:
* Update NGINX version to 1.15.6.

UPGRADE:
* For NGINX, use the 1.3.2 image from our DockerHub: `nginx/nginx-ingress:1.3.2` or `nginx/nginx-ingress:1.3.2-alpine`
* For NGINX Plus, please build your own image using the 1.3.2 source code.

### 1.3.1

CHANGES:
* Update NGINX Plus version to R15p2.

UPGRADE:
* For NGINX, use the 1.3.1 image from our DockerHub: `nginx/nginx-ingress:1.3.1` or `nginx/nginx-ingress:1.3.1-alpine`
* For NGINX Plus, please build your own image using the 1.3.1 source code.

### 1.3.0

IMPROVEMENTS:
* [325](https://github.com/nginxinc/kubernetes-ingress/pull/325): Report ingress status.
* [311](https://github.com/nginxinc/kubernetes-ingress/pull/311): Support JWT auth in mergeable minions.
* [310](https://github.com/nginxinc/kubernetes-ingress/pull/310): NGINX configuration template custom path support.
* [308](https://github.com/nginxinc/kubernetes-ingress/pull/308): Add prometheus exporter support to helm chart.
* [303](https://github.com/nginxinc/kubernetes-ingress/pull/303): Add fetch custom NGINX template from ConfigMap.
* [301](https://github.com/nginxinc/kubernetes-ingress/pull/301): Update prometheus exporter image for Plus.
* [298](https://github.com/nginxinc/kubernetes-ingress/pull/298): Prefetch ConfigMap before initial NGINX Config generation.
* [296](https://github.com/nginxinc/kubernetes-ingress/pull/296): Improve Helm Chart.
* [295](https://github.com/nginxinc/kubernetes-ingress/pull/295): Report version information.
* [294](https://github.com/nginxinc/kubernetes-ingress/pull/294): Support dynamic reconfiguration in mergeable ingresses for Plus.
* [287](https://github.com/nginxinc/kubernetes-ingress/pull/287): Support slow-start for Plus.
* [286](https://github.com/nginxinc/kubernetes-ingress/pull/286): Add support for active health checks for Plus.

CHANGES:
* [330](https://github.com/nginxinc/kubernetes-ingress/pull/330): Update NGINX version to 1.15.2.
* [329](https://github.com/nginxinc/kubernetes-ingress/pull/329): Enforce annotations inheritance in minions.

BUGFIXES:
* [326](https://github.com/nginxinc/kubernetes-ingress/pull/326): Fix find ingress for secret ns bug.
* [284](https://github.com/nginxinc/kubernetes-ingress/pull/284): Correct Logs for Mergeable Types with Duplicate Location. Thanks to [Fernando Diaz](https://github.com/diazjf).


UPGRADE:
* For NGINX, use the 1.3.0 image from our DockerHub: `nginx/nginx-ingress:1.3.0`
* For NGINX Plus, please build your own image using the 1.3.0 source code.

### 1.2.0

* [279](https://github.com/nginxinc/kubernetes-ingress/pull/279): Update dependencies.
* [278](https://github.com/nginxinc/kubernetes-ingress/pull/278): Fix mergeable Ingress types.
* [277](https://github.com/nginxinc/kubernetes-ingress/pull/277): Support grpc error responses.
* [276](https://github.com/nginxinc/kubernetes-ingress/pull/276): Add gRPC support.
* [274](https://github.com/nginxinc/kubernetes-ingress/pull/274): Change the default load balancing method to least_conn. 
* [272](https://github.com/nginxinc/kubernetes-ingress/pull/272): Move nginx-ingress image to the official nginx DockerHub.
* [268](https://github.com/nginxinc/kubernetes-ingress/pull/268): Correct Mergeable Types misspelling and optimize blacklists. Thanks to [Fernando Diaz](https://github.com/diazjf). 
* [266](https://github.com/nginxinc/kubernetes-ingress/pull/266): Add support for passive health checks.
* [261](https://github.com/nginxinc/kubernetes-ingress/pull/261): Update Customization Example.
* [258](https://github.com/nginxinc/kubernetes-ingress/pull/258): Handle annotations and conflicting paths for MergeableTypes. Thanks to [Fernando Diaz](https://github.com/diazjf).
* [256](https://github.com/nginxinc/kubernetes-ingress/pull/256): Add helm chart support. 
* [249](https://github.com/nginxinc/kubernetes-ingress/pull/249): Add support for prometheus for Plus.
* [241](https://github.com/nginxinc/kubernetes-ingress/pull/241): Update the doc about building the Docker image.
* [240](https://github.com/nginxinc/kubernetes-ingress/pull/240): Use new NGINX Plus API.
* [239](https://github.com/nginxinc/kubernetes-ingress/pull/239): Fix a typo in a variable name. Thanks to [Tony Li](https://github.com/mysterytony).
* [238](https://github.com/nginxinc/kubernetes-ingress/pull/238): Remove apt-get upgrade from Plus Dockerfile.
* [237](https://github.com/nginxinc/kubernetes-ingress/pull/237): Add unit test for ingress-class handling.
* [236](https://github.com/nginxinc/kubernetes-ingress/pull/236): Always respect `-ingress-class` option. Thanks to [Nick Novitski](https://github.com/nicknovitski).
* [235](https://github.com/nginxinc/kubernetes-ingress/pull/235): Change the base image to Debian Stretch for Plus controller.
* [234](https://github.com/nginxinc/kubernetes-ingress/pull/234): Update installation manifests and instructions.
* [233](https://github.com/nginxinc/kubernetes-ingress/pull/233): Add docker build options to Makefile.
* [231](https://github.com/nginxinc/kubernetes-ingress/pull/231): Prevent a possible failure of building Plus image.
* Documentation Fixes: [248](https://github.com/nginxinc/kubernetes-ingress/pull/248), thanks to [zariye](https://github.com/zariye). [252](https://github.com/nginxinc/kubernetes-ingress/pull/252). [270](https://github.com/nginxinc/kubernetes-ingress/pull/270).
* Update NGINX version to 1.13.12.
* Update NGINX Plus version to R15 P1.


### 1.1.1

* [228](https://github.com/nginxinc/kubernetes-ingress/pull/228): Add worker-rlimit-nofile configmap key. Thanks to [Aleksandr Lysenko](https://github.com/Sarga).
* [223](https://github.com/nginxinc/kubernetes-ingress/pull/223): Add worker-connections configmap key. Thanks to [Aleksandr Lysenko](https://github.com/Sarga).
* Update NGINX version to 1.13.8.

### 1.1.0

* [221](https://github.com/nginxinc/kubernetes-ingress/pull/221): Add git commit info to the IC log.
* [220](https://github.com/nginxinc/kubernetes-ingress/pull/220): Update dependencies.
* [213](https://github.com/nginxinc/kubernetes-ingress/pull/213): Add main snippets to allow Main context customization. Thanks to [Dewen Kong](https://github.com/kongdewen).
* [211](https://github.com/nginxinc/kubernetes-ingress/pull/211): Minimize the number of configuration reloads when the Ingress controller starts; fix a problem with endpoints updates for Plus.
* [208](https://github.com/nginxinc/kubernetes-ingress/pull/208): Add worker-shutdown-timeout configmap key. Thanks to [Aleksandr Lysenko](https://github.com/Sarga).
* [199](https://github.com/nginxinc/kubernetes-ingress/pull/199): Add support for Kubernetes ssl-redirect annotation. Thanks to [Luke Seelenbinder](https://github.com/lseelenbinder).
* [194](https://github.com/nginxinc/kubernetes-ingress/pull/194) Add keepalive configmap key and annotation.
* [193](https://github.com/nginxinc/kubernetes-ingress/pull/193): Add worker-cpu-affinity configmap key.
* [192](https://github.com/nginxinc/kubernetes-ingress/pull/192): Add worker-processes configmap key.
* [186](https://github.com/nginxinc/kubernetes-ingress/pull/186): Fix hardcoded controller class. Thanks to [Serhii M](https://github.com/SiriusRed).
* [184](https://github.com/nginxinc/kubernetes-ingress/pull/184): Return a meaningful error when there is no cert and key for the default server.
* Update NGINX version to 1.13.7.
* Makefile updates: golang container was updated to 1.9.

### 1.0.0

* [175](https://github.com/nginxinc/kubernetes-ingress/pull/175): Add support for JWT for NGINX Plus.
* [171](https://github.com/nginxinc/kubernetes-ingress/pull/171): Allow NGINX to listen on non-standard ports. Thanks to [Stanislav Seletskiy](https://github.com/seletskiy).
* [170](https://github.com/nginxinc/kubernetes-ingress/pull/170): Add the default server. **Note**: The Ingress controller will fail to start if there are no cert and key for the default server. You can pass a TLS Secret for the default server as an argument to the Ingress controller or add a cert and a key to the Docker image. 
* [169](https://github.com/nginxinc/kubernetes-ingress/pull/169): Ignore Ingress resources with empty hostnames.
* [168](https://github.com/nginxinc/kubernetes-ingress/pull/168): Add the `nginx.org/lb-method` annotation. Thanks to [Sajal Kayan](https://github.com/sajal).
* [166](https://github.com/nginxinc/kubernetes-ingress/pull/166): Watch Secret resources for updates. **Note**: If a Secret referenced by one or more Ingress resources becomes invalid or gets removed, the configuration for those Ingress resources will be disabled until there is a valid Secret.
* [160](https://github.com/nginxinc/kubernetes-ingress/pull/160): Add support for events. See the details [here](https://github.com/nginxinc/kubernetes-ingress/pull/160).
* [157](https://github.com/nginxinc/kubernetes-ingress/pull/157): Add graceful termination - when the Ingress controller receives `SIGTERM`, it shutdowns itself as well as NGINX, using `nginx -s quit`.

### 0.9.0

* [156](https://github.com/nginxinc/kubernetes-ingress/pull/156): Write a pem file with an SSL certificate and key atomically.
* [155](https://github.com/nginxinc/kubernetes-ingress/pull/155): Remove http2 annotation (http/2 can be enabled globally in the ConfigMap).
* [154](https://github.com/nginxinc/kubernetes-ingress/pull/154): Merge NGINX and NGINX Plus Ingress controller implementations.
* [151](https://github.com/nginxinc/kubernetes-ingress/pull/151): Use k8s.io/client-go.
* [146](https://github.com/nginxinc/kubernetes-ingress/pull/146): Fix health status.
* [141](https://github.com/nginxinc/kubernetes-ingress/pull/141): Set `worker_processes` to `auto` in NGINX configuration. Thanks to [Andreas Krüger](https://github.com/woopstar).
* [140](https://github.com/nginxinc/kubernetes-ingress/pull/140): Fix an error message. Thanks to [Andreas Krüger](https://github.com/woopstar).
* Update NGINX to version 1.13.3.

### 0.8.1

* Update NGINX version to 1.13.0.

### 0.8.0

* [117](https://github.com/nginxinc/kubernetes-ingress/pull/117): Add a customization option: location-snippets, server-snippets and http-snippets. Thanks to [rchicoli](https://github.com/rchicoli).
* [116](https://github.com/nginxinc/kubernetes-ingress/pull/116): Add support for the 301 redirect to https based on the `http_x_forwarded_proto` header. Thanks to [Chris](https://github.com/cwhenderson20).
* Update NGINX version to 1.11.13.
* Makefile updates: gcloud docker push command; golang container was updated to 1.8.
* Documentation fixes: [113](https://github.com/nginxinc/kubernetes-ingress/pull/113). Thanks to [Linus Lewandowski](https://github.com/LEW21).

### 0.7.0

* [108](https://github.com/nginxinc/kubernetes-ingress/pull/108): Support for the `server_tokens` directive via the annotation and in the configmap. Thanks to [David Radcliffe](https://github.com/dwradcliffe).
* [103](https://github.com/nginxinc/kubernetes-ingress/pull/103): Improve error reporting when NGINX fails to start.
* [100](https://github.com/nginxinc/kubernetes-ingress/pull/100): Add the health check location. Thanks to [Julian](https://github.com/jmastr).
* [95](https://github.com/nginxinc/kubernetes-ingress/pull/95): Fix the runtime.TypeAssertionError issue, which sometimes occurred when deleting resources. Thanks to [Tang Le](https://github.com/tangle329).
* [93](https://github.com/nginxinc/kubernetes-ingress/pull/93): Fix overwriting of Secrets with the same name from different namespaces.
* [92](https://github.com/nginxinc/kubernetes-ingress/pull/92/files): Add overwriting of the HSTS header. Previously, when HSTS was enabled, if a backend issued the HSTS header, the controller would add the second HSTS header. Now the controller overwrites the HSTS header, if a backend also issues it.
* [91](https://github.com/nginxinc/kubernetes-ingress/pull/91):
Fix the issue with single service Ingress resources without any Ingress rules: the controller didn't pick up any updates of the endpoints of the service of such an Ingress resource. Thanks to [Tang Le](https://github.com/tangle329).
* [88](https://github.com/nginxinc/kubernetes-ingress/pull/88): Support for the `proxy_hide_header` and the `proxy_pass_header` directives via annotations and in the configmap. Thanks to [Nico Schieder](https://github.com/thetechnick).
* [85](https://github.com/nginxinc/kubernetes-ingress/pull/85): Add the configmap settings to support perfect forward secrecy. Thanks to [Nico Schieder](https://github.com/thetechnick).
* [84](https://github.com/nginxinc/kubernetes-ingress/pull/84): Secret retry: If a certificate Secret referenced in an Ingress object is not found,
the Ingress controller will reject the Ingress object. but retries every 5s. Thanks to [Nico Schieder](https://github.com/thetechnick).
* [81](https://github.com/nginxinc/kubernetes-ingress/pull/81): Add configmap options to turn on the PROXY protocol. Thanks to [Nico Schieder](https://github.com/thetechnick).
* Update NGINX version to 1.11.8.
* Documentation fixes: [104](https://github.com/nginxinc/kubernetes-ingress/pull/104/files) and [97](https://github.com/nginxinc/kubernetes-ingress/pull/97/files). Thanks to [Ruilin Huang](https://github.com/hrl) and [Justin Garrison](https://github.com/rothgar).

### 0.6.0

* [75](https://github.com/nginxinc/kubernetes-ingress/pull/75): Add the HSTS settings in the configmap and annotations. Thanks to [Nico Schieder](https://github.com/thetechnick).
* [74](https://github.com/nginxinc/kubernetes-ingress/pull/74): Fix the issue of the `kubernetes.io/ingress.class` annotation handling. Thanks to [Tang Le](https://github.com/tangle329).
* [70](https://github.com/nginxinc/kubernetes-ingress/pull/70): Add support for the alpine-based image for the NGINX controller.
* [68](https://github.com/nginxinc/kubernetes-ingress/pull/68): Support for proxy-buffering settings in the configmap and annotations. Thanks to [Mark Daniel Reidel](https://github.com/df-mreidel).
* [66](https://github.com/nginxinc/kubernetes-ingress/pull/66): Support for custom log-format in the configmap. Thanks to [Mark Daniel Reidel](https://github.com/df-mreidel).
* [65](https://github.com/nginxinc/kubernetes-ingress/pull/65): Add HTTP/2 as an option in the configmap and annotations. Thanks to [Nico Schieder](https://github.com/thetechnick).
* The NGINX Plus controller image is now based on Ubuntu Xenial.

### 0.5.0

* Update NGINX version to 1.11.5.
* [64](https://github.com/nginxinc/kubernetes-ingress/pull/64): Add the `nginx.org/rewrites` annotation, which allows to rewrite the URI of a request before sending it to the application. Thanks to [Julian](https://github.com/jmastr).
* [62](https://github.com/nginxinc/kubernetes-ingress/pull/62): Add the `nginx.org/ssl-services` annotation, which allows load balancing of HTTPS applications. Thanks to [Julian](https://github.com/jmastr).

### 0.4.0

* [54](https://github.com/nginxinc/kubernetes-ingress/pull/54): Previously, when specifying the port of a service in an Ingress rule, you had to use the value of the target port of that port of the service, which was incorrect. Now you must use the port value or the name of the port of the service instead of the target port value. **Note**: Please make necessary changes to your Ingress resources, if ports of your services have different values of the port and the target port fields.
* [55](https://github.com/nginxinc/kubernetes-ingress/pull/55): Add support for the `kubernetes.io/ingress.class` annotation in Ingress resources.
* [58](https://github.com/nginxinc/kubernetes-ingress/pull/58): Add the version information to the controller. For each version of the NGINX controller, you can find a corresponding image on [DockerHub](https://hub.docker.com/r/nginxdemos/nginx-ingress/tags/) with a tag equal to the version. The latest version is available through the `latest` tag.

The previous version was 0.3


### Notes

* Except when mentioned otherwise, the controller refers both to the NGINX and the NGINX Plus Ingress controllers.
