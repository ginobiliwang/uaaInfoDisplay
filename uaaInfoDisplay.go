package main

import (
       "github.com/bitly/go-simplejson"
	   "github.com/modood/table"
	   "io/ioutil"
	   "strings"
	   "bytes"
       "fmt"
	   "os"
	   "os/exec"
	   "bufio"
	   "strconv"
	   "errors"
	   "flag"
)

type UserInfo struct {
	UserName string
	Origin string
	IsPksClustersAdmin bool
	IsPksClustersManage bool
	//BindingName string
}

type Binding struct {
	UserName string
	BindingName string
	SubjectKind string
	RoleKind string
	RoleName string
	RoleDesc []string
}

func fetchUserInfoFromUaa(uaaApi string, uaaClientSecret string) {
    //fmt.Printf("Info: uaac fetch user informations\n")
	exeCmd := "uaac target " + uaaApi + ":8443 --skip-ssl-validation;uaac token client get admin -s  " + uaaClientSecret  +"; uaac users | sed 's/\\<meta\\>/meta\\:/g'  | sed 's/\\<name\\>/name\\:/g'  | yq . > ./input.json"
	CommandWrapperWithoutReturn(exeCmd)
}

func CommandWrapperWithoutReturn (userCmd string) {
	//fmt.Printf("command is \n %s \n", userCmd)
	cmd := exec.Command("/bin/bash", "-c", userCmd)

    userOut, err := cmd.StdoutPipe()
    if err != nil {
        fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
        return
    }

    if err := cmd.Start(); err != nil {
        fmt.Println("Error:The command is err,", err)
        return
    }
	
	//bytes, err := ioutil.ReadAll(userOut)
	_, err = ioutil.ReadAll(userOut)
    if err != nil {
		fmt.Println("ReadAll Stdout:", err.Error())
		return
	}
	//fmt.Printf("stdout:\n\n %s", bytes)

    if err := cmd.Wait(); err != nil {
        fmt.Println("wait:", err.Error())
        return
    }
}

func CommandWrapperWithStringReturn (userCmd string) string {
	
	result := ""
	//fmt.Printf("command is \n %s \n", userCmd)
	cmd := exec.Command("/bin/bash", "-c", userCmd)

    userOut, err := cmd.StdoutPipe()
    if err != nil {
        fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
        return result
    }

    if err := cmd.Start(); err != nil {
        fmt.Println("Error:The command is err,", err)
        return result
    }
	
	bytes, err := ioutil.ReadAll(userOut)
    if err != nil {
		fmt.Println("ReadAll Stdout:", err.Error())
		return result
	}
	//fmt.Printf("stdout:\n\n %s", bytes)
	result = string(bytes)

    if err := cmd.Wait(); err != nil {
        fmt.Println("wait:", err.Error())
        return result
    }
	return result
}

func CommandWrapperWithArrayReturn (userCmd string) ([]string) {
	
	result 	:= []string{}
	//fmt.Printf("command is \n %s \n", userCmd)
	cmd := exec.Command("/bin/bash", "-c", userCmd)

    userOut, err := cmd.StdoutPipe()
    if err != nil {
        fmt.Printf("Error:can not obtain stdout pipe for command:%s\n", err)
        return result
    }

    if err := cmd.Start(); err != nil {
        fmt.Println("Error:The command is err,", err)
        return result
    }

	
	bytes, err := ioutil.ReadAll(userOut)
    if err != nil {
		fmt.Println("ReadAll Stdout:", err.Error())
		return result
	}
	//fmt.Printf("stdout:\n\n %s", bytes)
	s := string(bytes)

    if err := cmd.Wait(); err != nil {
        fmt.Println("wait:", err.Error())
        return result
    }
	
	for _, lineStr := range strings.Split(s, "\n") {
        lineStr = strings.TrimSpace(lineStr)
        if lineStr == "" {
            continue
        }
        result = append(result, lineStr)
    	//fmt.Println("cluster name is: %s\n", lineStr)
	}
	return result
}

func PickupOneCluster(clusterPool []string) string {
	fmt.Printf("Please choose the cluster: \n")
	for i, _ := range clusterPool {
			fmt.Printf("%d) %s\n", i, clusterPool[i])
	}

	inputReader := bufio.NewReader(os.Stdin)
	input, err  := inputReader.ReadString('\n')
	if err == nil {
		fmt.Printf("The input was: %s", input)
	}
	index, _ := strconv.Atoi(strings.TrimSuffix(input, "\n"))
	//fmt.Printf("The input was: %d", index)
	return clusterPool[index]
}

func fetchBindingInfoFromKubeApi(operatorUserName string,  operatorPassword string, uaaApi string) {
	// 1) pks login using alana
	// 2) alana get-credentials using the cluster name from first cluster (if there is no cluster, just exit)
	// 3) alana using kubectl get clusterrolebinding and rolebinding, put into binding.json
	//fmt.Printf("Info: fetch information from kubeApi\n")
	exeCmd := "pks logout >/dev/null  2>&1;pks login -a " + uaaApi + " -u " + operatorUserName + " -p " + operatorPassword  + " -k > /dev/null  2>&1; pks clusters 2>&1  | grep succeeded   | awk '{print $1}'"
	clusterPool := CommandWrapperWithArrayReturn(exeCmd)
	clusterName := PickupOneCluster(clusterPool)
	//fmt.Printf("clusterName is : %s\n", clusterName)
	exeCmd = "echo " + operatorPassword + " | pks get-credentials " + clusterName 
	CommandWrapperWithoutReturn(exeCmd)
	exeCmd = "kubectl config use-context " + clusterName 
	CommandWrapperWithoutReturn(exeCmd)
	exeCmd = "kubectl get clusterrolebindings -l generated=true -o json > ./clusterrolebinding.json"
	CommandWrapperWithoutReturn(exeCmd)
	exeCmd = "kubectl get rolebindings -o json > ./rolebinding.json"
	CommandWrapperWithoutReturn(exeCmd)
    // the following step should be carried out in main function
    // 4) parse binding.json, get subject, map User information back (with ClusterRole & Role)
	
}


func parseUserInfo () (map[string]UserInfo, error) {
    //fmt.Printf("Info: json information parsing from input.json\n")
	var userInfoMap map[string]UserInfo
	userInfoMap = make(map[string]UserInfo)
	
	doc, _ := ioutil.ReadFile("./input.json")
	dec, _ := simplejson.NewFromReader(bytes.NewReader(doc))
	userArray, _ := dec.Get("resources").Array()	
	//fmt.Println(len(userArray))

	if (0 == len(userArray)) {
		//fmt.Printf("Fatal: no user information available and exit\n")
		//return
		return nil,errors.New("Fatal: no user information available and exit")
	}
	
	var userName, origin string
	var isPksClustersAdmin bool
	var isPksClustersManage bool
	var userInfoUnit UserInfo

	for i, _ := range userArray {
		user 	 := dec.Get("resources").GetIndex(i)
		userName = user.Get("username").MustString()
		origin	 = user.Get("origin").MustString()

	    isPksClustersManage = false
		isPksClustersAdmin  = false

		groupArray, _ := user.Get("groups").Array()
		//fmt.Println(len(groupArray))
		for j, _ := range groupArray {
			oneGroup	:= user.Get("groups").GetIndex(j)
			gName 		:= oneGroup.Get("display").MustString()
			if ( 0 == strings.Compare(gName, "pks.clusters.admin")){
				isPksClustersAdmin = true
			} else if (0 == strings.Compare(gName, "pks.clusters.manage")){
				isPksClustersManage = true	
			}
		}

		//store information in userInfoMap

		userInfoUnit.UserName 				= userName
		userInfoUnit.Origin 				= origin
		userInfoUnit.IsPksClustersAdmin 	= isPksClustersAdmin
		userInfoUnit.IsPksClustersManage 	= isPksClustersManage
		userInfoMap[userName]				= userInfoUnit
		
		//fmt.Printf("username=%s, Origin=%s, IsPksClustersAdmin=%t, IsPksClustersAdmin=%t\n", UserName, Origin, IsPksClustersAdmin, IsPksClustersManage)
	}
	return userInfoMap, nil
}

func parseBinding (bindingKind string) (map[string]Binding, error) {
    
	//fmt.Printf("Info: json information parsing from clusterrolebinding.json and rolebinding.json \n")

	var bindingMap map[string]Binding
	bindingMap = make(map[string]Binding)
	var binding Binding
	var bindingName string
	var subjectKind string
	var roleKind string
	//we reuse UserName from UserInfo map struct
	var roleName string
	var roleDesc []string
	var exeCmd string
	var tmpString string
	var doc []byte
	var userName string

	if (0 == strings.Compare("clusterRoleBinding", bindingKind)) {
		doc, _ = ioutil.ReadFile("./clusterrolebinding.json")
	} else if (0 == strings.Compare("roleBinding", bindingKind)) {
		doc, _ = ioutil.ReadFile("./rolebinding.json")
	} else {
		return nil, errors.New("My Error Message!")
	}
	
	dec, _ := simplejson.NewFromReader(bytes.NewReader(doc))
	clusterRoleBindingArray, _ := dec.Get("items").Array()	
	//fmt.Println(len(clusterRoleBindingArray))

	if (0 == len(clusterRoleBindingArray)) {
		//fmt.Printf("Fatal: no clusterrole binding available and exit\n")
		return nil, errors.New("no binding to show!")
	}

	for i, _ := range clusterRoleBindingArray {
		oneBinding 	 	  := dec.Get("items").GetIndex(i)
		subjectsArray, _  := oneBinding.Get("subjects").Array()	
		//fmt.Println(len(subjectsArray))
		
		for j, _ := range subjectsArray {
			subject			 :=  oneBinding.Get("subjects").GetIndex(j)
			subjectKind	  	 =	subject.Get("kind").MustString()  
			userName 		 = 	subject.Get("name").MustString()
			//fmt.Println("we get subject kind: %s \n", SubjectKind)
			if ( 0 == strings.Compare(subjectKind, "User")) {
				bindingName 	= oneBinding.Get("metadata").Get("name").MustString()
				roleKind		= oneBinding.Get("roleRef").Get("kind").MustString()  
				roleName		= oneBinding.Get("roleRef").Get("name").MustString()

				exeCmd 		= "kubectl get " + roleKind + " " + roleName + " " + "--template={{.rules}}"			
				tmpString 	= CommandWrapperWithStringReturn(exeCmd)
				//fmt.Println("exeCmd to get verbs: \n %s \n", exeCmd)
				if (0 == strings.Compare(tmpString, "")) {
					continue
				} else {
					tmpString = strings.TrimPrefix(tmpString, "[map")
					tmpString = strings.TrimSuffix(tmpString, "]")
					roleDesc  = strings.Split(tmpString, " map")
				}
			} else {
				continue
			}
			//store information in bindingMap
			binding.UserName					= userName
			binding.BindingName					= bindingName
			binding.SubjectKind					= subjectKind
			binding.RoleKind					= roleKind
			binding.RoleName					= roleName
			binding.RoleDesc					= roleDesc
			bindingMap[userName]				= binding
		}
	}

	//traverse the map to print out
	/*
	for k, v := range userInfoMap {
		fmt.Printf("username=%s, username=%s, Origin=%s, IsPksClustersAdmin=%t, IsPksClustersManage=%t\n",k, v.UserName, v.Origin, v.IsPksClustersAdmin, v.IsPksClustersManage)
	}
	*/
	return bindingMap, nil
}

func printSlice(x []UserInfo){
   fmt.Printf("len=%d cap=%d slice=%v\n",len(x),cap(x),x)
}


func drawUserInfoTable (userInfoM map[string]UserInfo) (error) {
	if (0 == len(userInfoM)) {
		return errors.New("do not draw userTable\n")
	}

	var s []UserInfo = make([]UserInfo, 0)
	for _, v := range userInfoM {
		s = append(s, v)
	}
	t := table.AsciiTable(s)
	//t := table.Table(s)
	fmt.Println(t)
	return nil	
}

func drawBindingTable (BindingM map[string]Binding) (error) {
	if (0 == len(BindingM)) {
		return errors.New("do not draw bindingTable\n")
	}
	
	var s []Binding = make([]Binding, 0)
	for _, v := range BindingM {
		s = append(s, v)
	}
	//t := table.AsciiTable(s)
	t := table.Table(s)
	fmt.Println(t)
	return nil	
}

func main() {

	uaaApi				:= flag.String("api", "empty", "what is the uaaApi url?")
	uaaClientSecret		:= flag.String("key", "empty", "what is the uaaClientSecret?")
	operatorUserName	:= flag.String("op", "empty", "what is the operatorUserName? alana?")
	operatorPassword	:= flag.String("passwd", "empty", "what is the operatorPassword?")

	flag.Parse()
	//others := flag.Args()

	//fmt.Println("uaaApi: ", *uaaApi)
    //fmt.Println("uaaClientSecret: ", *uaaClientSecret)
    //fmt.Println("operatorUserName: ", *operatorUserName)
    //fmt.Println("operatorPassword: ", *operatorPassword)
    //fmt.Println("others: ", others)

	fetchUserInfoFromUaa(*uaaApi, *uaaClientSecret)
	fetchBindingInfoFromKubeApi(*operatorUserName, *operatorPassword, *uaaApi)
	userInfoMap, err := parseUserInfo()
	if (err != nil) {
		fmt.Printf("error happen when parsing userInfo\n")
	}

	roleBindingMap, err := parseBinding("roleBinding")
	if (err != nil) {
		fmt.Printf("error happen when parsing roleBinding\n")
	}

	clusterRoleBindingMap, err := parseBinding("clusterRoleBinding")
	if (err != nil) {
		fmt.Printf("error happen when parsing clusterRoleBinding\n")
	}

	fmt.Printf("\n\n")
	err = drawUserInfoTable(userInfoMap)
	if (err != nil) {
		fmt.Printf("error happen when draw userInfo Table\n")
	}

	err = drawBindingTable(clusterRoleBindingMap)
	if (err != nil) {
		fmt.Printf("error happen when draw clsuterRoleBinding Table\n")
	}

	err = drawBindingTable(roleBindingMap)
	if (err != nil) {
		fmt.Printf("error happen when draw roleBinding Table\n")
	}

	/*
	fmt.Printf("\n\n")
	
	for k, v := range userInfoMap {
		fmt.Printf("username=%s, username=%s, Origin=%s, IsPksClustersAdmin=%t, IsPksClustersManage=%t\n",k, v.UserName, v.Origin, v.IsPksClustersAdmin, v.IsPksClustersManage)
	}
	fmt.Printf("\n\n")
	
	for k, v := range clusterRoleBindingMap {
		fmt.Printf("username=%s, username=%s, BindingName=%s, SubjectKind=%s, RoleKind=%s, RoleName=%s, RoleDesc=%q\n",k, v.UserName, v.BindingName, v.SubjectKind, v.RoleKind, v.RoleName, v.RoleDesc)
	}
	fmt.Printf("\n\n")

	for k, v := range roleBindingMap {
		fmt.Printf("username=%s, username=%s, BindingName=%s, SubjectKind=%s, RoleKind=%s, RoleName=%s, RoleDesc=%q\n",k, v.UserName, v.BindingName, v.SubjectKind, v.RoleKind, v.RoleName, v.RoleDesc)
	}
	*/

}
