# Contributing

## Contribute Workflow

### Pull requests

All the repositories accept contributions via [GitHub Pull requests (PR)](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/about-pull-requests).

PRs are always welcome, even if they only contain small fixes like typos or a few lines of code.

### Fork and clone

Fork the Lunettes repository and clone the code to your local workspace. 

```sh
$ dir="$GOPATH/src/github.com/alipay" 
$ mkdir -p "$dir"
$ cd "$dir"
$ git clone https://github.com/{your-github-username}/container-observability-service
$ cd container-observability-service
```

#### Configure the upstream remote

Next, add the remote `upstream`. Configuring this remote allows you to
synchronize your forked copy, `origin`, with the `upstream`. 

```sh
$ git remote add upstream https://github.com/alipay/container-observability-service
```

Run `git remote -v`. Your remotes should appear similar to these:

```sh
origin  https://github.com/your-github-username/container-observability-service.git (fetch)  
origin  https://github.com/your-github-username/container-observability-service.git (push)  
upstream  https://github.com/alipay/container-observability-service (fetch)  
upstream  https://github.com/alipay/container-observability-service (push)  
```

### GitHub workflow

#### Create a topic branch

Create a new "topic branch" to do your work on:

```sh
$ git checkout -b fix-contrib-bugs
```

You should *always* create a new "topic branch" for PR work.

Then you can make your changes.

### Develop, Build and Test

Write code on the new branch in your fork and test it locally.

### Create PR

#### Commit your code

Commit your changes to the current (`fix-contrib-bugs`) branch:

```sh
$ git commit -s -m 'This is my commit message'
```

**NOTE**: you must use the `-s` to sign your commit to pass the [DCO](https://developercertificate.org/) check.

#### Push code and create PR

Push your local `fix-contrib-bugs` branch to your remote fork:

```sh
$ git push -u origin fix-contrib-bugs
```

Create the PR:
  - Browse to https://github.com/alipay/container-observability-service.
  - Click the "Compare & pull request" button that appears.
  - Click the "Create pull request" button.

###  Keep sync with upstream

Once your branch gets out of sync with the upstream branch, use the following commands to update:

```sh
$ git checkout main
$ git fetch upstream
$ git rebase upstream/main
```
