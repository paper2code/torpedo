package steam

import (
	"torpedobot/common"
	"fmt"
	"golang.org/x/net/html"
	"bytes"
	"strings"
	"net/url"
	"strconv"
)


const StoreURL = "http://store.steampowered.com/explore/new/"


type Client struct {
	StoreURL string
}


type GameItem struct {
	GameURL string
	GameThumbnail string
	CurrentPrice float64
	RegularPrice float64
	DiscountPercentage int
	Platforms []string
}


func (gi *GameItem) IsComplete() (result bool) {
	if gi.GameURL != "" && gi.GameThumbnail != "" {
		result = true
	}
	return
}


func parsedURL(full_url string) (result string) {
	parsed, _ := url.Parse(full_url)
	result = fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, parsed.Path)
	return
}


func (cli *Client) ShowNew() (items []*GameItem){

	data, err := common.GetURLBytes(cli.StoreURL)

	if err != nil {
		fmt.Printf("Fetching URL %s failed with %+v\n", cli.StoreURL, err)
	}

	result, err := html.Parse(bytes.NewReader(data))

	if err != nil {
		fmt.Printf("Parsing URL %s failed with %+v\n", cli.StoreURL, err)
	}

	var f func(*html.Node)

	ourlink := false

	f = func(n *html.Node) {
		// <a href="">.... </a>
		if n.Type == html.ElementNode && n.Data == "a" {
			ourlink = false
			game_item := &GameItem{}
			for _, attr := range n.Attr {
				if attr.Key == "class" && strings.Trim(attr.Val, " ") == "tab_item" {
					ourlink = true
				}
				if attr.Key == "href" {
					game_item.GameURL = parsedURL(attr.Val)
				}
			}
			if ourlink {
				for mc := n.FirstChild; mc != nil; mc = mc.NextSibling {
					// <a href="game_url"> kids
					// <div><img..></div>
					if mc.Data == "div" && len(mc.Attr) > 0 && mc.Attr[0].Val == "tab_item_cap" {
						for zc := mc.FirstChild; zc != nil; zc = zc.NextSibling {
							if zc.Data != "img" || len(zc.Attr) != 2{
								continue
							}
 							game_item.GameThumbnail = zc.Attr[1].Val
							break
						}
					}

					// <div class="discount_block..">
					if mc.Data == "div" && len(mc.Attr) > 0 && strings.HasPrefix(mc.Attr[0].Val, "discount_block") {

						for yc := mc.FirstChild; yc != nil; yc = yc.NextSibling {

							// <div class="discount_pct
							if yc.Data == "div" && len(yc.Attr) >0 && yc.Attr[0].Val == "discount_pct" {
								price := strings.TrimRight(yc.FirstChild.Data, "%")
								result, err := strconv.Atoi(price)
								if err == nil {
									game_item.DiscountPercentage = result
								}
							}

							// <div class="discount_prices">
							for zc := yc.FirstChild; zc != nil; zc = zc.NextSibling {
								if len(zc.Attr) == 0 || yc.Attr[0].Val != "discount_prices" {
									continue
								}
								switch zc.Attr[0].Val {
								case "discount_original_price":
									price := strings.TrimLeft(zc.FirstChild.Data, "$")
									result, err := strconv.ParseFloat(price, 64)
									if err == nil {
										game_item.RegularPrice = result
									}
								case "discount_final_price":
									price := strings.TrimLeft(zc.FirstChild.Data, "$")
									result, err := strconv.ParseFloat(price, 64)
									if err == nil {
										game_item.CurrentPrice = result
									}
								}

							}
						}
					}

					// <div class="tab_item_content">
					if mc.Data == "div" && len(mc.Attr) > 0 && mc.Attr[0].Val == "tab_item_content" {
						for zc := mc.FirstChild; zc != nil; zc = zc.NextSibling {
							if len(zc.Attr) == 0 || zc.Attr[0].Val != "tab_item_details" {
								continue
							}
							for dt := zc.FirstChild; dt != nil; dt = dt.NextSibling {
								if len(dt.Attr) == 0 || ! strings.HasPrefix(dt.Attr[0].Val, "platform_img") {
									continue
								}
								platform :=  strings.TrimSpace(strings.TrimLeft(dt.Attr[0].Val, "platform_img"))
								if platform == "hmd_separator" {
									continue
								}
								game_item.Platforms = append(game_item.Platforms, platform)
							}
						}
					}
				}
				if game_item.IsComplete() {
					items = append(items, game_item)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(result)
	return
}


func  NewClient() (client *Client){
	client = &Client{}
	client.StoreURL = StoreURL
	return
}