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
)

var token_variable = ""

const keychain_app_service = "github-forkrefresh"
const username = "dmore"

func store_secret_on_keychain(token string ){

	service := keychain_app_service
    user := username
    password := token
    //if you want to inject it onto or from an env var...
    //password := token
    //os.Setenv("GITHUB_TOKEN",password)
    //password = envVariable("GITHUB_TOKEN")
  
    // set password
    err := keyring.Set(service, user, password)
    if err != nil {
        log.Fatal(err)
    }
}

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
	//store_secret_on_keychain("GITHUB_TOKEN_WITH_RIGHTS")
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
		fmt.Println("reponame: %v", reponame)
		
		//LOOP HERE EACH ELEMENT
	    var ret = ""
	    //master call
		ret,err = fork_refresh_call("master", reponame, "POST")
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
			//check this call for 'Not Found'
			if strings.Contains(string(ret), "Not Found") {
				fmt.Println("Not Found found on stringifed response => [develop]")
				//develop call
				ret,err = fork_refresh_call("develop", reponame, "POST")
				if err != nil {
					log.Fatalln(err)
					//ret2,err2 := call("", "main", "", "POST")
					//if err2 != nil {
					//	log.Fatalln(err)
					//	//continue
					//}
				}
				//os.Exit(0)
				//break 
			}

		}else{
			fmt.Println("ok")
			//continue
		}
		
	}
	//exit after looping repo names
	os.Exit(0)

}


func fork_refresh_call(branch string, reponame string, method string) (string, error) {
	absPath, _ := filepath.Abs("../"+ branch + ".json")
	f, err := os.Open(absPath)
	if err != nil {
	    log.Fatal(err)
	}
	defer f.Close()

	reponame = strings.TrimSuffix(reponame, "/")
	reponame = strings.TrimPrefix(reponame, "/")
	
	httpposturl := "https://api.github.com/repos/" + reponame + "/merge-upstream"
	fmt.Println("url: %v", httpposturl)
	request, err := http.NewRequest("POST", httpposturl, f)
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
	//fmt.Println("response :", response.Errorf)
	fmt.Println("response Status:", response.Status)
	b, err := io.ReadAll(response.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
		return "nil", err
	}
	return string(b), nil
	//return fmt.Println(string(b))
}
