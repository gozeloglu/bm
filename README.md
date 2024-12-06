# bm

New version of the [bm.sh](https://github.com/gozeloglu/bm.sh) tool written in Go.

## Installation

You can type the following command.

```shell
go install github.com/gozeloglu/bm@latest
```

### Local development build

You can build and run the application with the following command. It can be used for local development tests.
```shell
go build bm.go &&  go install bm.go
```

## Usage

Currently, limited commands are provided.

### Save new bookmark

```shell
bm --save
```
![img_1.png](docs/img/img_1.png)

After saving the link:

![img.png](docs/img/img.png)

### List all bookmarks

```shell
bm
bm --list # this is another option for listing
```
You can navigate the links with up and down arrow keys.

![img_2.png](docs/img/img_2.png)

Type `/` for searching a specific bookmarked link.

![img_3.png](docs/img/img_3.png)

### Delete the bookmark

```shell
bm --delete
```

Just use **backspace** to delete.

### Version

```shell
bm --version
```