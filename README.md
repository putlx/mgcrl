## Usage

```
Usage: mgcrl [options] URL
Options:
  -c string
        volumes or chapters (default "1:-1")
  -m int
        max retry time (default 5)
  -o string
        output directory (default ".")
  -v string
        manga version
```

## Example

Execute `mgcrl -c=1,-2:-1 http://www.700mh.com/manhua/1436` to download the first one and the last two chapters of 「The Promised Neverland」.

Execute `mgcrl -v=单行本 -c=1:第03卷 https://www.mhgui.com/comic/4683/` to download the first three volumes of 「D.Gray-man」.

## Supported Websites

| Website | Example |
| ------- | -------- |
| dmzj | https://www.dmzj.com/info/biaoren.html<br>https://www.dmzj1.com/info/biaoren.html<br>https://manhua.dmzj.com/tianlaizhishengdetianshi<br>https://manhua.dmzj1.com/tianlaizhishengdetianshi |
| katui | http://www.700mh.com/manhua/1436 |
| laimanhua | https://www.laimanhua.com/kanmanhua/33952/ |
| lsj | https://lsj.ac/comic/xugoutuile |
| mangadex | https://mangadex.org/title/50810/natsu-no-mamono `-v group` |
| manhua123 | https://m.manhua123.net/comic/8199.html |
| manhuadb | https://www.manhuadb.com/manhua/1011 |
| manhuagui | https://www.mhgui.com/comic/991/ `-v`<br>https://www.manhuagui.com/comic/842 `-v`<br>https://tw.manhuagui.com/comic/842 `-v` |
| tieba | https://tieba.baidu.com/p/6910932313 |

## License

GPL-3.0
