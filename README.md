# bm-go

New version of the [bm.sh](https://github.com/gozeloglu/bm.sh) tool written in Go.

## Installation

You can type the following command.

```shell
go install github.com/gozeloglu/bm@latest
```

```shell
go build cmd/* &&  go install cmd/bm.go cmd/cmd.go
```

## Usage

Currently, limited commands are provided.

### Save new bookmark

```shell
bm --save https://google.com
```

### List all bookmarks

```shell
bm --list
```

### Delete the bookmark

```shell
bm --delete 3 # deletes the 3rd bookmark
```

### Update the bookmark link

```shell
bm --update 3 https://github.com  # updates the 3rd link with new one, github.com in that example
```

### Open link on the browser

```shell
bm --open 3 # opens 3rd link on the default browser
```

### Export links

```shell
bm --export ~/links/  # exports links to ~/links directory as .db 
```

### Version

```shell
bm --version
```