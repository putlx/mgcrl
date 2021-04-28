package ext

import (
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/putlx/mgcrl/util"
)

func Manhuagui(URL, ver string) (m Manga, err error) {
	html, err := util.GetHtml(URL)
	if err != nil {
		return
	}
	m.Title = html.Find("div.book-title h1").Text()
	m.Delay = 1500 * time.Millisecond

	host := regexp.MustCompile(`(www\.manhuagui|tw\.manhuagui|www\.mhgui)\.com`).FindString(URL)
	html.Find("h4").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if len(ver) == 0 || s.Find("span").Text() == ver {
			for s = s.Next(); !strings.Contains(s.AttrOr("class", ""), "chapter-list"); {
				s = s.Next()
			}
			s.Find("ul").Each(func(_ int, s *goquery.Selection) {
				var cs []Chapter
				s.Find("li a").Each(func(_ int, s *goquery.Selection) {
					href, exists := s.Attr("href")
					if !exists {
						panic(`no attribute "href"`)
					}
					cs = append(cs, Chapter{"https://" + host + href, s.Find("span").Contents().First().Text()})
				})
				m.Chapters = append(cs, m.Chapters...)
			})
			return len(ver) == 0
		}
		return true
	})
	util.Reverse(m.Chapters)
	return
}

func ManhuaguiImages(URL string) ([]string, error) {
	text, err := util.GetText(URL)
	if err != nil {
		return nil, err
	}
	script := regexp.MustCompile(`window\[".+?"\](.+?)\</`).FindStringSubmatch(text)[1]
	script = util.EvalJavaScript(
		`var LZString=(function(){var f=String.fromCharCode;var keyStrBase64="ABCDEFG` +
			`HIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=";var baseReverseD` +
			`ic={};function getBaseValue(alphabet,character){if(!baseReverseDic[alphabet]` +
			`){baseReverseDic[alphabet]={};for(var i=0;i<alphabet.length;i++){baseReverse` +
			`Dic[alphabet][alphabet.charAt(i)]=i}}return baseReverseDic[alphabet][charact` +
			`er]}var LZString={decompressFromBase64:function(input){if(input==null)return` +
			`"";if(input=="")return null;return LZString._0(input.length,32,function(inde` +
			`x){return getBaseValue(keyStrBase64,input.charAt(index))})},_0:function(leng` +
			`th,resetValue,getNextValue){var dictionary=[],next,enlargeIn=4,dictSize=4,nu` +
			`mBits=3,entry="",result=[],i,w,bits,resb,maxpower,power,c,data={val:getNextV` +
			`alue(0),position:resetValue,index:1};for(i=0;i<3;i+=1){dictionary[i]=i}bits=` +
			`0;maxpower=Math.pow(2,2);power=1;while(power!=maxpower){resb=data.val&data.p` +
			`osition;data.position>>=1;if(data.position==0){data.position=resetValue;data` +
			`.val=getNextValue(data.index++)}bits|=(resb>0?1:0)*power;power<<=1}switch(ne` +
			`xt=bits){case 0:bits=0;maxpower=Math.pow(2,8);power=1;while(power!=maxpower)` +
			`{resb=data.val&data.position;data.position>>=1;if(data.position==0){data.pos` +
			`ition=resetValue;data.val=getNextValue(data.index++)}bits|=(resb>0?1:0)*powe` +
			`r;power<<=1}c=f(bits);break;case 1:bits=0;maxpower=Math.pow(2,16);power=1;wh` +
			`ile(power!=maxpower){resb=data.val&data.position;data.position>>=1;if(data.p` +
			`osition==0){data.position=resetValue;data.val=getNextValue(data.index++)}bit` +
			`s|=(resb>0?1:0)*power;power<<=1}c=f(bits);break;case 2:return""}dictionary[3` +
			`]=c;w=c;result.push(c);while(true){if(data.index>length){return""}bits=0;max` +
			`power=Math.pow(2,numBits);power=1;while(power!=maxpower){resb=data.val&data.` +
			`position;data.position>>=1;if(data.position==0){data.position=resetValue;dat` +
			`a.val=getNextValue(data.index++)}bits|=(resb>0?1:0)*power;power<<=1}switch(c` +
			`=bits){case 0:bits=0;maxpower=Math.pow(2,8);power=1;while(power!=maxpower){r` +
			`esb=data.val&data.position;data.position>>=1;if(data.position==0){data.posit` +
			`ion=resetValue;data.val=getNextValue(data.index++)}bits|=(resb>0?1:0)*power;` +
			`power<<=1}dictionary[dictSize++]=f(bits);c=dictSize-1;enlargeIn--;break;case` +
			` 1:bits=0;maxpower=Math.pow(2,16);power=1;while(power!=maxpower){resb=data.v` +
			`al&data.position;data.position>>=1;if(data.position==0){data.position=resetV` +
			`alue;data.val=getNextValue(data.index++)}bits|=(resb>0?1:0)*power;power<<=1}` +
			`dictionary[dictSize++]=f(bits);c=dictSize-1;enlargeIn--;break;case 2:return ` +
			`result.join("")}if(enlargeIn==0){enlargeIn=Math.pow(2,numBits);numBits++}if(` +
			`dictionary[c]){entry=dictionary[c]}else{if(c===dictSize){entry=w+w.charAt(0)` +
			`}else{return null}}result.push(entry);dictionary[dictSize++]=w+entry.charAt(` +
			`0);enlargeIn--;w=entry;if(enlargeIn==0){enlargeIn=Math.pow(2,numBits);numBit` +
			`s++}}}};return LZString})();String.prototype.splic=function(f){return LZStri` +
			`ng.decompressFromBase64(this).split(f)};` + script)

	path := regexp.MustCompile(`"path":"(.+?)"`).FindStringSubmatch(script)[1]
	m := regexp.MustCompile(`"m":"(.+?)"`).FindStringSubmatch(script)[1]
	e := regexp.MustCompile(`"e":(\d+)`).FindStringSubmatch(script)[1]
	script = regexp.MustCompile(`\[(.+?)\]`).FindStringSubmatch(script)[1]
	mat := regexp.MustCompile(`"(.+?)"`).FindAllStringSubmatch(script, -1)
	imgs := make([]string, len(mat))
	for i := range mat {
		imgs[i] = "https://i.hamreus.com" + path + mat[i][1] + "?e=" + e + "&m=" + m
	}
	return imgs, nil
}
