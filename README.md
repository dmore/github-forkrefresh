README.md

github-forkrefresh
==================

Very basic app that Will refresh the oiginal project from your public forks so they are updated. I need something like this because I have many forks and want all of them updated so I know what's going on the parent projects. Tested this with hundred or so repos and it was ok.

- Needs GITHUB_TOKEN

- Stores/retrieves token on your OS/Keychain.

- Needs a list of the public repos you want to keep updated from your original projects.


What does it do:
===============

    It calls github api to discover what branch is used on the forking. Uses that to post it to the fork refresh so your public forks are up-to-date with its source and ought to trigger a remote refresh of them branches.
    
    there is a repos_repo.json json array file. make sure your forking public repos are there.
    That is your forks, not the originals.

    repos_repo.json
    [
       "yourgithubuser/yourpublicfork",
       "yourgithubuser/yourpublicfork2"
    ]

    tells github to refresh the fork from the original so your public forks are refreshed from the source.


What does it need:
==================


    a) it needs your public fork repos as above. youruser/yourpublicfork

    b|) it also uses go-keyring to pull the GITHUB_TOKEN secret from the OS/Keychain
    so you'll need to store the token in the OS/keychain first and retrieve it from there.

    Otherwise feel free to change the code and use an env var instead. That code is commented out.

    c) THe token needs to have quite a bit of rights to keep GITHUB happy so keep that in mind.

    d) you can inject it on a line commmented out. 

    func main() {

    //uncomment to store your secret o keychain
    //store_secret_on_keychain("GITHUB_TOKEN_WITH_RIGHTS")
    token_variable = retrieve_secret_from_keychain()

Dependencies:
=============
    Depends on zalando/go-keyring to retrieve and pull secrets. Currently using version 0.2.3.
    I would have used a different approach but for this this is fine.

    Also uses Jeffail/gabs to construct the expected json at the remote end.

How to run it:
==============
    cd github-forkrefresh/httpclient
    go run main.go

    this is the core of it, if you just wanna know

```go
func fork_refresh_call(branch string, reponame string, method string) (string, error) {
    //now that we know the branch name in advance we can use that instead of this.
    
    jsonObj := gabs.New()
    // or gabs.Wrap(jsonObject) to work on an existing map[string]interface{}

    //jsonObj.Set("branch", "" + branch)
    jsonObj.Set("" + branch, "branch")

    jsonOutput := jsonObj.String()

    fmt.Println(jsonObj.String())
    fmt.Println(jsonObj.StringIndent("", "  "))

    var jsonStr = []byte(jsonOutput)
    //req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

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
    // b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
    if err != nil {
        log.Fatalln(err)
        return "nil", err
    }

    //fmt.Println("response :", response.Errorf)
    fmt.Println("response Status:", response.Status)
    fmt.Println("response Body:", string(b))
    return string(b), nil
    //return fmt.Println(string(b))
}


```