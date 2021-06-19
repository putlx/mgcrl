## Usage

```
Usage: mgcrl get <URL> [options]
       mgcrl serve <PORT> [options]

Options for get:
  -c string
        volumes or chapters (default "1:-1")
  -m int
        max retry time (default 3)
  -o string
        output directory (default ".")
  -v string
        manga version

Options for serve:
  -f string
        auto crawl manga according to the config file
  -l string
        redirect log to file
```

Option `-c` is followed by argument with the format of `from[:to][,from[:to]]...` where `from` and `to` can both be chapter title or index.

## Example

Execute `mgcrl get http://www.700mh.com/manhua/1436 -c=1,-2:-1` to download the first one and the last two chapters of 「The Promised Neverland」.

Execute `mgcrl get https://www.mhgui.com/comic/4683/ -v=单行本 -c=1:第03卷` to download the first three volumes of 「D.Gray-man」.

Execute `mgcrl serve <PORT>` and open `http://localhost:<PORT>/` in your browser to access WebUI.

For automated crawling, take `config.json` as an example.

## Supported Websites

| Website | Example |
| ------- | -------- |
| dmzj | https://www.dmzj.com/info/biaoren.html<br>https://www.dmzj1.com/info/biaoren.html<br>https://manhua.dmzj.com/tianlaizhishengdetianshi<br>https://manhua.dmzj1.com/tianlaizhishengdetianshi |
| fzdm | https://manhua.fzdm.com/2/ |
| guoguomh | https://mh.guoguomh.com/manhua/shiyouzhiguo/ |
| katui | http://www.700mh.com/manhua/1436 |
| laimanhua | https://www.laimanhua.com/kanmanhua/33952/ |
| lsj | https://lsj.ac/comic/xugoutuile |
| manhua123 | https://m.manhua123.net/comic/8199.html |
| manhuadb | https://www.manhuadb.com/manhua/1011 |
| manhuagui | https://www.mhgui.com/comic/991/ `-v`<br>https://www.manhuagui.com/comic/842 `-v`<br>https://tw.manhuagui.com/comic/842 `-v` |
| mm1316 | https://www.mm1316.com/rexue/BluePeriod/ |
| tieba | https://tieba.baidu.com/p/6910932313 |

## License

GPL-3.0
