package main

import (
	"fmt"
	"net/http"
	"os"
	//"time"
	"io/ioutil"
	"io"
	"log"
	"path/filepath"
	"encoding/json"
	"strings"
	"github.com/zalando/go-keyring"
	"github.com/Jeffail/gabs"
	"bytes"
)

var token_variable = ""

const keychain_app_service = "github-forkrefresh"
const username = "dmore"

//this method stores your secret on the OS keychain. 
func store_secret_on_keychain(token string ){

	service := keychain_app_service
    user := username
    password := token
    //if you want to inject it onto or from an env var...
    //password := token
    //os.Setenv("GITHUB_TOKEN",password)
    //password = os.Getenv("GITHUB_TOKEN")
  
    // set password
    err := keyring.Set(service, user, password)
    if err != nil {
        log.Fatal(err)
    }
}

//this method retrieves secret from the keychain. 
func retrieve_secret_from_keychain() (string){

	service := keychain_app_service
    user := username

	// get password
    secret, err := keyring.Get(service, user)
    if err != nil {
        log.Fatal(err)
    }

    log.Println(secret)
    return secret
}


func main() {

	//uncomment to store your secret o keychain
	//store_secret_on_keychain("GITHUB_TOKEN")
	token_variable = retrieve_secret_from_keychain()

	//file must be json array not json
	absPath, _ := filepath.Abs("../repos_repo.json")
	f, err := os.Open(absPath)
	if err != nil {
	    log.Fatal(err)
	}
	fmt.Println("Successfully Opened repos_repo.json")
	defer f.Close()

	//unmarshall
	byteValue, _ := ioutil.ReadAll(f)   
	var arr []string
	//var result map[string]interface{} not working. 
	//in memory test works fine also
	/**
	var dataJson = `[
    	"dmore/aws-vault-local-os-keychain-mfa",
    	"dmore/aws-multi-region-cicd-with-terraform"
	]`
	err2 := json.Unmarshal([]byte(dataJson), &arr)
	if err2 != nil {
      fmt.Println("error2:", err2)
      os.Exit(1)
    }
    **/
	
	err3 := json.Unmarshal([]byte(byteValue), &arr)
	if err3 != nil {
      fmt.Println("error3:", err3)
      os.Exit(1)
    }
    log.Printf("Unmarshaled: %v", arr)
    //loop through
    for i := 0; i < len(arr); i++ {
		
		var reponame = string(arr[i])
		reponame = strings.TrimSuffix(reponame, "/")
		reponame = strings.TrimPrefix(reponame, "/")
		fmt.Println(reponame)
		
		//LOOP HERE EACH ELEMENT

	    var ret = ""
	    var branch = ""
	    //grab branches first so we know what branch name we need in advance...
		branch, err = fork_get_query_branch(reponame)
		if err != nil {
			log.Fatalln(err)
			//ret2,err2 := call("", "main", "", "POST")
			//if err2 != nil {
			//	log.Fatalln(err)
			//	//continue
			//}
		}
	    //relevant branch call
		ret,err = fork_refresh_call(branch, reponame, "POST")
		if err != nil {
			log.Fatalln(err)
			//ret2,err2 := call("", "main", "", "POST")
			//if err2 != nil {
			//	log.Fatalln(err)
			//	//continue
			//}
		}
		//don't print it too much content
		//fmt.Println(string(ret))
		
		if strings.Contains(string(ret), "Not Found") {
			fmt.Println("checking...")
			fmt.Println(string(ret))

			fmt.Println("Not Found found on stringifed response => [main]")
			//main call
			ret,err = fork_refresh_call("main", reponame, "POST")
			if err != nil {
				log.Fatalln(err)
				//ret2,err2 := call("", "main", "", "POST")
				//if err2 != nil {
				//	log.Fatalln(err)
				//	//continue
				//}
			}

		}else{
			fmt.Println("ok")
			//continue
		}
		
	}
	//exit after looping repo names
	os.Exit(0)

}


func fork_get_query_branch(reponame string) (string, error) {

	reponame = strings.TrimSuffix(reponame, "/")
	reponame = strings.TrimPrefix(reponame, "/")

	httpposturl := "https://api.github.com/repos/" + reponame + "/branches"
	fmt.Println("url: %v", httpposturl)
	request, err := http.NewRequest("GET", httpposturl, nil)
	if err != nil {
	    log.Fatal(err)
	}
	//request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Accept", "Accept: application/vnd.github+json")
	request.Header.Set("Authorization", "token " + token_variable)
	request.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
	    log.Fatal(err)
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
		return "nil", err
	}

	//fmt.Println("response :", response.Errorf)
	fmt.Println("response Status:", response.Status)
	fmt.Println("response Body:", string(b))


	var objmap []map[string]interface{}
	if err := json.Unmarshal(b, &objmap); err != nil {
    	log.Fatal(err)
	}
	fmt.Println(objmap[0]["name"])

	var return_branch = ""
	for k, v := range objmap[0] {
	    switch c := v.(type) {
	    case string:
	    	if k == "name" {
	    		return_branch = string(c)
	    		fmt.Printf("Item %q is a string, containing %q\n", k, c)
	    	}
	        
	    case float64:
	        //fmt.Printf("Looks like item %q is a number, specifically %f\n", k, c)
	        continue
	    default:
	        //fmt.Printf("Not sure what type item %q is, but I think it might be %T\n", k, c)
	        continue
	    }
	}
	fmt.Println("return_branch is " + return_branch)
	return string(return_branch), nil
}

func fork_refresh_call(branch string, reponame string, method string) (string, error) {
	//now that we know the branch name in advance we can use that instead of this.
	
	jsonObj := gabs.New()
	// or gabs.Wrap(jsonObject) to work on an existing map[string]interface{}

	jsonObj.Set("" + branch, "branch")

	jsonOutput := jsonObj.String()

	fmt.Println(jsonObj.String())
	fmt.Println(jsonObj.StringIndent("", "  "))

	var jsonStr = []byte(jsonOutput)
   
	reponame = strings.TrimSuffix(reponame, "/")
	reponame = strings.TrimPrefix(reponame, "/")

	httpposturl := "https://api.github.com/repos/" + reponame + "/merge-upstream"
	fmt.Println("url: %v", httpposturl)
	request, err := http.NewRequest("POST", httpposturl, bytes.NewBuffer(jsonStr))
	if err != nil {
	    log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	request.Header.Set("Accept", "application/vnd.github.v3+json")
	request.Header.Set("Authorization", "token " + token_variable)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
	    log.Fatal(err)
	}
	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalln(err)
		return "nil", err
	}

	fmt.Println("response Status:", response.Status)
	fmt.Println("response Body:", string(b))
	return string(b), nil
	//return fmt.Println(string(b))
}
