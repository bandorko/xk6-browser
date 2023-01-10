package browser

import (
	"fmt"

	"github.com/dop251/goja"

	"github.com/grafana/xk6-browser/api"
	"github.com/grafana/xk6-browser/chromium"

	k6common "go.k6.io/k6/js/common"
	k6modules "go.k6.io/k6/js/modules"
)

// mapping is a type of mapping between our API (api/) and the JS
// module. It acts like a bridge and allows adding wildcard methods
// and customization over our API.
//
// TODO
// We should put this type back in when the following issue is resolved
// on the Goja side:
// https://github.com/dop251/goja/issues/469
//
// type mapping map[string]any

// wildcards is a list of extra mappings for our API (api/).
func wildcards() map[string]string {
	return map[string]string{
		"Page.query":             "$",
		"Page.queryAll":          "$$",
		"Frame.query":            "$",
		"Frame.queryAll":         "$$",
		"ElementHandle.query":    "$",
		"ElementHandle.queryAll": "$$",
	}
}

// mapBrowserToGoja maps the browser API to the JS module.
// The motivation of this mapping was to support $ and $$ wildcard
// methods.
// See issue #661 for more details.
func mapBrowserToGoja(vu k6modules.VU) *goja.Object {
	rt := vu.Runtime()

	var (
		obj         = rt.NewObject()
		browserType = chromium.NewBrowserType(vu)
	)
	for k, v := range mapBrowserType(rt, browserType) {
		err := obj.Set(k, rt.ToValue(v))
		if err != nil {
			k6common.Throw(rt, fmt.Errorf("mapping: %w", err))
		}
	}

	return obj
}

// mapRequest to the JS module.
func mapRequest(rt *goja.Runtime, r api.Request) map[string]any {
	maps := map[string]any{
		"allHeaders": r.AllHeaders,
		"failure":    r.Failure,
		"frame": func() *goja.Object {
			mf := mapFrame(rt, r.Frame())
			return rt.ToValue(mf).ToObject(rt)
		},
		"headerValue":         r.HeaderValue,
		"headers":             r.Headers,
		"headersArray":        r.HeadersArray,
		"isNavigationRequest": r.IsNavigationRequest,
		"method":              r.Method,
		"postData":            r.PostData,
		"postDataBuffer":      r.PostDataBuffer,
		"postDataJSON":        r.PostDataJSON,
		"redirectedFrom": func() *goja.Object {
			mr := mapRequest(rt, r.RedirectedFrom())
			return rt.ToValue(mr).ToObject(rt)
		},
		"redirectedTo": func() *goja.Object {
			mr := mapRequest(rt, r.RedirectedTo())
			return rt.ToValue(mr).ToObject(rt)
		},
		"resourceType": r.ResourceType,
		"response": func() *goja.Object {
			mr := mapResponse(rt, r.Response())
			return rt.ToValue(mr).ToObject(rt)
		},
		"size":   r.Size,
		"timing": r.Timing,
		"url":    r.URL,
	}

	return maps
}

// mapResponse to the JS module.
func mapResponse(rt *goja.Runtime, r api.Response) map[string]any {
	maps := map[string]any{
		"allHeaders": r.AllHeaders,
		"body":       r.Body,
		"finished":   r.Finished,
		"frame": func() *goja.Object {
			mf := mapFrame(rt, r.Frame())
			return rt.ToValue(mf).ToObject(rt)
		},
		"headerValue":  r.HeaderValue,
		"headerValues": r.HeaderValues,
		"headers":      r.Headers,
		"headersArray": r.HeadersArray,
		"jSON":         r.JSON,
		"ok":           r.Ok,
		"request": func() *goja.Object {
			mr := mapRequest(rt, r.Request())
			return rt.ToValue(mr).ToObject(rt)
		},
		"securityDetails": r.SecurityDetails,
		"serverAddr":      r.ServerAddr,
		"size":            r.Size,
		"status":          r.Status,
		"statusText":      r.StatusText,
		"url":             r.URL,
	}

	return maps
}

// mapElementHandle to the JS module.
//
//nolint:funlen
func mapElementHandle(rt *goja.Runtime, eh api.ElementHandle) map[string]any {
	maps := map[string]any{
		"asElement": func() *goja.Object {
			m := mapElementHandle(rt, eh.AsElement())
			return rt.ToValue(m).ToObject(rt)
		},
		"dispose":        eh.Dispose,
		"evaluate":       eh.Evaluate,
		"evaluateHandle": eh.EvaluateHandle,
		"getProperties":  eh.GetProperties,
		"getProperty":    eh.GetProperty,
		"jSONValue":      eh.JSONValue,
		"objectID":       eh.ObjectID,
		"boundingBox":    eh.BoundingBox,
		"check":          eh.Check,
		"click":          eh.Click,
		"contentFrame": func() *goja.Object {
			f := eh.ContentFrame()
			mf := mapFrame(rt, f)
			return rt.ToValue(mf).ToObject(rt)
		},
		"dblclick":      eh.Dblclick,
		"dispatchEvent": eh.DispatchEvent,
		"fill":          eh.Fill,
		"focus":         eh.Focus,
		"getAttribute":  eh.GetAttribute,
		"hover":         eh.Hover,
		"innerHTML":     eh.InnerHTML,
		"innerText":     eh.InnerText,
		"inputValue":    eh.InputValue,
		"isChecked":     eh.IsChecked,
		"isDisabled":    eh.IsDisabled,
		"isEditable":    eh.IsEditable,
		"isEnabled":     eh.IsEnabled,
		"isHidden":      eh.IsHidden,
		"isVisible":     eh.IsVisible,
		"ownerFrame": func() *goja.Object {
			f := eh.OwnerFrame()
			mf := mapFrame(rt, f)
			return rt.ToValue(mf).ToObject(rt)
		},
		"press":                  eh.Press,
		"screenshot":             eh.Screenshot,
		"scrollIntoViewIfNeeded": eh.ScrollIntoViewIfNeeded,
		"selectOption":           eh.SelectOption,
		"selectText":             eh.SelectText,
		"setInputFiles":          eh.SetInputFiles,
		"tap":                    eh.Tap,
		"textContent":            eh.TextContent,
		"type":                   eh.Type,
		"uncheck":                eh.Uncheck,
		"waitForElementState":    eh.WaitForElementState,
		"waitForSelector": func(selector string, opts goja.Value) *goja.Object {
			eh := eh.WaitForSelector(selector, opts)
			ehm := mapElementHandle(rt, eh)
			return rt.ToValue(ehm).ToObject(rt)
		},
	}
	maps["$"] = func(selector string) *goja.Object {
		eh := eh.Query(selector)
		ehm := mapElementHandle(rt, eh)
		return rt.ToValue(ehm).ToObject(rt)
	}
	maps["$$"] = func(selector string) *goja.Object {
		var (
			mehs []map[string]any
			ehs  = eh.QueryAll(selector)
		)
		for _, eh := range ehs {
			ehm := mapElementHandle(rt, eh)
			mehs = append(mehs, ehm)
		}
		return rt.ToValue(mehs).ToObject(rt)
	}

	return maps
}

// mapFrame to the JS module.
//
//nolint:funlen
func mapFrame(rt *goja.Runtime, f api.Frame) map[string]any {
	maps := map[string]any{
		"addScriptTag": f.AddScriptTag,
		"addStyleTag":  f.AddStyleTag,
		"check":        f.Check,
		"childFrames": func() *goja.Object {
			var (
				mcfs []map[string]any
				cfs  = f.ChildFrames()
			)
			for _, fr := range cfs {
				mcfs = append(mcfs, mapFrame(rt, fr))
			}
			return rt.ToValue(mcfs).ToObject(rt)
		},
		"click":          f.Click,
		"content":        f.Content,
		"dblclick":       f.Dblclick,
		"dispatchEvent":  f.DispatchEvent,
		"evaluate":       f.Evaluate,
		"evaluateHandle": f.EvaluateHandle,
		"fill":           f.Fill,
		"focus":          f.Focus,
		"frameElement": func() *goja.Object {
			eh := mapElementHandle(rt, f.FrameElement())
			return rt.ToValue(eh).ToObject(rt)
		},
		"getAttribute": f.GetAttribute,
		"goto":         f.Goto,
		"hover":        f.Hover,
		"innerHTML":    f.InnerHTML,
		"innerText":    f.InnerText,
		"inputValue":   f.InputValue,
		"isChecked":    f.IsChecked,
		"isDetached":   f.IsDetached,
		"isDisabled":   f.IsDisabled,
		"isEditable":   f.IsEditable,
		"isEnabled":    f.IsEnabled,
		"isHidden":     f.IsHidden,
		"isVisible":    f.IsVisible,
		"iD":           f.ID,
		"loaderID":     f.LoaderID,
		"locator":      f.Locator,
		"name":         f.Name,
		"page": func() *goja.Object {
			mp := mapPage(rt, f.Page())
			return rt.ToValue(mp).ToObject(rt)
		},
		"parentFrame": func() *goja.Object {
			mf := mapFrame(rt, f.ParentFrame())
			return rt.ToValue(mf).ToObject(rt)
		},
		"press":             f.Press,
		"selectOption":      f.SelectOption,
		"setContent":        f.SetContent,
		"setInputFiles":     f.SetInputFiles,
		"tap":               f.Tap,
		"textContent":       f.TextContent,
		"title":             f.Title,
		"type":              f.Type,
		"uncheck":           f.Uncheck,
		"url":               f.URL,
		"waitForFunction":   f.WaitForFunction,
		"waitForLoadState":  f.WaitForLoadState,
		"waitForNavigation": f.WaitForNavigation,
		"waitForSelector": func(selector string, opts goja.Value) *goja.Object {
			eh := f.WaitForSelector(selector, opts)
			ehm := mapElementHandle(rt, eh)
			return rt.ToValue(ehm).ToObject(rt)
		},
		"waitForTimeout": f.WaitForTimeout,
	}
	maps["$"] = func(selector string) *goja.Object {
		eh := f.Query(selector)
		ehm := mapElementHandle(rt, eh)
		return rt.ToValue(ehm).ToObject(rt)
	}
	maps["$$"] = func(selector string) *goja.Object {
		var (
			mehs []map[string]any
			ehs  = f.QueryAll(selector)
		)
		for _, eh := range ehs {
			ehm := mapElementHandle(rt, eh)
			mehs = append(mehs, ehm)
		}
		return rt.ToValue(mehs).ToObject(rt)
	}

	return maps
}

// mapPage to the JS module.
//
//nolint:funlen
func mapPage(rt *goja.Runtime, p api.Page) map[string]any {
	maps := map[string]any{
		"addInitScript":           p.AddInitScript,
		"addScriptTag":            p.AddScriptTag,
		"addStyleTag":             p.AddStyleTag,
		"bringToFront":            p.BringToFront,
		"check":                   p.Check,
		"click":                   p.Click,
		"close":                   p.Close,
		"content":                 p.Content,
		"context":                 p.Context,
		"dblclick":                p.Dblclick,
		"dispatchEvent":           p.DispatchEvent,
		"dragAndDrop":             p.DragAndDrop,
		"emulateMedia":            p.EmulateMedia,
		"emulateVisionDeficiency": p.EmulateVisionDeficiency,
		"evaluate":                p.Evaluate,
		"evaluateHandle":          p.EvaluateHandle,
		"exposeBinding":           p.ExposeBinding,
		"exposeFunction":          p.ExposeFunction,
		"fill":                    p.Fill,
		"focus":                   p.Focus,
		"frame":                   p.Frame,
		"frames": func() *goja.Object {
			var (
				mfrs []map[string]any
				frs  = p.Frames()
			)
			for _, fr := range frs {
				mfrs = append(mfrs, mapFrame(rt, fr))
			}
			return rt.ToValue(mfrs).ToObject(rt)
		},
		"getAttribute": p.GetAttribute,
		"goBack":       p.GoBack,
		"goForward":    p.GoForward,
		"goto":         p.Goto,
		"hover":        p.Hover,
		"innerHTML":    p.InnerHTML,
		"innerText":    p.InnerText,
		"inputValue":   p.InputValue,
		"isChecked":    p.IsChecked,
		"isClosed":     p.IsClosed,
		"isDisabled":   p.IsDisabled,
		"isEditable":   p.IsEditable,
		"isEnabled":    p.IsEnabled,
		"isHidden":     p.IsHidden,
		"isVisible":    p.IsVisible,
		"locator":      p.Locator,
		"mainFrame": func() *goja.Object {
			mf := mapFrame(rt, p.MainFrame())
			return rt.ToValue(mf).ToObject(rt)
		},
		"opener": p.Opener,
		"pause":  p.Pause,
		"pdf":    p.Pdf,
		"press":  p.Press,
		"reload": func(opts goja.Value) *goja.Object {
			r := mapResponse(rt, p.Reload(opts))
			return rt.ToValue(r).ToObject(rt)
		},
		"route":                       p.Route,
		"screenshot":                  p.Screenshot,
		"selectOption":                p.SelectOption,
		"setContent":                  p.SetContent,
		"setDefaultNavigationTimeout": p.SetDefaultNavigationTimeout,
		"setDefaultTimeout":           p.SetDefaultTimeout,
		"setExtraHTTPHeaders":         p.SetExtraHTTPHeaders,
		"setInputFiles":               p.SetInputFiles,
		"setViewportSize":             p.SetViewportSize,
		"tap":                         p.Tap,
		"textContent":                 p.TextContent,
		"title":                       p.Title,
		"type":                        p.Type,
		"uncheck":                     p.Uncheck,
		"unroute":                     p.Unroute,
		"url":                         p.URL,
		"video":                       p.Video,
		"viewportSize":                p.ViewportSize,
		"waitForEvent":                p.WaitForEvent,
		"waitForFunction":             p.WaitForFunction,
		"waitForLoadState":            p.WaitForLoadState,
		"waitForNavigation":           p.WaitForNavigation,
		"waitForRequest":              p.WaitForRequest,
		"waitForResponse":             p.WaitForResponse,
		"waitForSelector":             p.WaitForSelector,
		"waitForTimeout":              p.WaitForTimeout,
		"workers":                     p.Workers,
	}
	maps["$"] = func(selector string) *goja.Object {
		eh := p.Query(selector)
		ehm := mapElementHandle(rt, eh)
		return rt.ToValue(ehm).ToObject(rt)
	}
	maps["$$"] = func(selector string) *goja.Object {
		var (
			mehs []map[string]any
			ehs  = p.QueryAll(selector)
		)
		for _, eh := range ehs {
			ehm := mapElementHandle(rt, eh)
			mehs = append(mehs, ehm)
		}
		return rt.ToValue(mehs).ToObject(rt)
	}

	return maps
}

// mapBrowserContext to the JS module.
func mapBrowserContext(rt *goja.Runtime, bc api.BrowserContext) map[string]any {
	return map[string]any{
		"addCookies":                  bc.AddCookies,
		"addInitScript":               bc.AddInitScript,
		"browser":                     bc.Browser,
		"clearCookies":                bc.ClearCookies,
		"clearPermissions":            bc.ClearPermissions,
		"close":                       bc.Close,
		"cookies":                     bc.Cookies,
		"exposeBinding":               bc.ExposeBinding,
		"exposeFunction":              bc.ExposeFunction,
		"grantPermissions":            bc.GrantPermissions,
		"newCDPSession":               bc.NewCDPSession,
		"route":                       bc.Route,
		"setDefaultNavigationTimeout": bc.SetDefaultNavigationTimeout,
		"setDefaultTimeout":           bc.SetDefaultTimeout,
		"setExtraHTTPHeaders":         bc.SetExtraHTTPHeaders,
		"setGeolocation":              bc.SetGeolocation,
		"setHTTPCredentials":          bc.SetHTTPCredentials, //nolint:staticcheck
		"setOffline":                  bc.SetOffline,
		"storageState":                bc.StorageState,
		"unroute":                     bc.Unroute,
		"waitForEvent":                bc.WaitForEvent,
		"pages": func() *goja.Object {
			var (
				mpages []map[string]any
				pages  = bc.Pages()
			)
			for _, page := range pages {
				if page == nil {
					continue
				}
				m := mapPage(rt, page)
				mpages = append(mpages, m)
			}

			return rt.ToValue(mpages).ToObject(rt)
		},
		"newPage": func() *goja.Object {
			page := bc.NewPage()
			m := mapPage(rt, page)
			return rt.ToValue(m).ToObject(rt)
		},
	}
}

// mapBrowser to the JS module.
func mapBrowser(rt *goja.Runtime, b api.Browser) map[string]any {
	return map[string]any{
		"close":       b.Close,
		"contexts":    b.Contexts,
		"isConnected": b.IsConnected,
		"on":          b.On,
		"userAgent":   b.UserAgent,
		"version":     b.Version,
		"newContext": func(opts goja.Value) *goja.Object {
			bctx := b.NewContext(opts)
			m := mapBrowserContext(rt, bctx)
			return rt.ToValue(m).ToObject(rt)
		},
		"newPage": func(opts goja.Value) *goja.Object {
			page := b.NewPage(opts)
			m := mapPage(rt, page)
			return rt.ToValue(m).ToObject(rt)
		},
	}
}

// mapBrowserType to the JS module.
func mapBrowserType(rt *goja.Runtime, bt api.BrowserType) map[string]any {
	return map[string]any{
		"connect":                 bt.Connect,
		"executablePath":          bt.ExecutablePath,
		"launchPersistentContext": bt.LaunchPersistentContext,
		"name":                    bt.Name,
		"launch": func(opts goja.Value) *goja.Object {
			m := mapBrowser(rt, bt.Launch(opts))
			return rt.ToValue(m).ToObject(rt)
		},
	}
}