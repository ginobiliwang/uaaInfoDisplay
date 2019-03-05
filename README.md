## Features

This tool could only be used for PKS(Pivotal Container Service)
https://docs.pivotal.io/runtimes/pks/1-2/index.html

* This code piece is running on go language 1.11.x 
* It will retrive user information from UAA server.
* It will retrive clusterrole binding information from k8s.
* It will retrive role binding information from k8s.
* It will draw three tables with above three pieces of informations.


## Install

You will need to install uaac ruby client side from your side and make sure it is working.
https://github.com/cloudfoundry/cf-uaac

You will need to install accordingly pks and kubectl binary.
https://network.pivotal.io/

```bash
go get github.com/modood/table
go get github.com/bitly/go-simplejson
go get github.com/ginobiliwang/uaaInfoDisplay
```
You will need to make sure that there is an operator ahead of time. If not, please create a new operator.
https://docs.pivotal.io/runtimes/pks/1-2/manage-users.html
```bash
uaac target https://PKS-API:8443 --skip-ssl-validation
uaac token client get admin -s ADMIN-CLIENT-SECRET
uaac user add alana --emails alana@example.com -p password
uaac member add pks.clusters.admin alana
uaac member add pks.clusters.manage alana
```

## Usage

```bash
Usage of uaaInfoDisplay:
  -api string
    	what is the uaaApi url? (default "empty")
  -key string
    	what is the uaaClientSecret? (default "empty")
  -op string
    	what is the operatorUserName? alana? (default "empty")
  -passwd string
    	what is the operatorPassword? (default "empty")
```
```bash
uaaInfoDisplay  -api pks-23.haas-59.pez.pivotal.io -key Mb_1rO1wx-DXzA4I7M9kDhGHd37s4G_r  -op alana -passwd changeme
Please choose the cluster:
0) test2
1) david-cluster
2) test
2
The input was: 2


+----------+--------+--------------------+---------------------+
| UserName | Origin | IsPksClustersAdmin | IsPksClustersManage |
+----------+--------+--------------------+---------------------+
| usera    | ldap   | true               | true                |
| wangfg   | uaa    | false              | true                |
| admin    | uaa    | true               | false               |
| alana    | uaa    | true               | false               |
| dzhou    | uaa    | true               | false               |
+----------+--------+--------------------+---------------------+
┌──────────┬─────────────────────┬─────────────┬─────────────┬───────────────┬───────────────────────────────────────────────────────────────────────────┐
│ UserName │ BindingName         │ SubjectKind │ RoleKind    │ RoleName      │ RoleDesc                                                                  │
├──────────┼─────────────────────┼─────────────┼─────────────┼───────────────┼───────────────────────────────────────────────────────────────────────────┤
│ alana    │ alana-cluster-admin │ User        │ ClusterRole │ cluster-admin │ [[apiGroups:[*] resources:[*] verbs:[*]] [nonResourceURLs:[*] verbs:[*]]] │
│ usera    │ usera-cluster-admin │ User        │ ClusterRole │ cluster-admin │ [[apiGroups:[*] resources:[*] verbs:[*]] [nonResourceURLs:[*] verbs:[*]]] │
└──────────┴─────────────────────┴─────────────┴─────────────┴───────────────┴───────────────────────────────────────────────────────────────────────────┘
┌──────────┬─────────────┬─────────────┬──────────┬────────────┬───────────────────────────────────────────────┐
│ UserName │ BindingName │ SubjectKind │ RoleKind │ RoleName   │ RoleDesc                                      │
├──────────┼─────────────┼─────────────┼──────────┼────────────┼───────────────────────────────────────────────┤
│ wangfg   │ read-pods   │ User        │ Role     │ pod-reader │ [[apiGroups:[] resources:[pods] verbs:[get]]] │
└──────────┴─────────────┴─────────────┴──────────┴────────────┴───────────────────────────────────────────────┘
```

## License

GNU General Public License
