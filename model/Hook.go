package databath

import (
	"log"

	"github.com/daemonl/go_gsd/components"
)

type IHook interface {
	Applies(*components.HookContext) bool
	RunPreHook(*components.HookContext)
	RunPostHook(*components.HookContext)
}

type Hook struct {
	Collection string                 `json:"collection"`
	When       HookWhen               `json:"when"`
	Set        map[string]interface{} `json:"set"`
	Email      *HookEmail             `json:"email"`
	Raw        []string               `json:"raw"`
	Scripts    []string               `json:"scripts"`
}

type HookWhen struct {
	Field string `json:"field"`
	What  string `json:"what"`
}

type HookEmail struct {
	Recipient string `json:"recipient"`
	Template  string `json:"template"`
}

func (hook *Hook) Applies(hc *components.HookContext) bool {
	if hook.Collection != hc.ActionSummary.Collection {
		return false
	}
	if hook.When.What != hc.ActionSummary.Action {
		return false
	}
	if len(hook.When.Field) > 0 {
		if _, ok := hc.ActionSummary.Fields[hook.When.Field]; !ok {
			return false
		}
	}
	return true
}

func (hook *Hook) RunPreHook(hc *components.HookContext) {
	// Look, I'll be honest, this probably doesn't work...
	for k, v := range hook.Set {
		_, exists := hc.ActionSummary.Fields[k]
		if exists {
			continue
		}
		vString, ok := v.(string)
		if ok {
			if vString == "#me" {
				v = hc.ActionSummary.UserId
			}
		}
		hc.ActionSummary.Fields[k] = v
	}
}

func (hook *Hook) RunPostHook(hc *components.HookContext) {

	for _, raw := range hook.Raw {
		hc.DB.Exec(raw)
	}
	for _, scriptName := range hook.Scripts {

		log.Printf("Hook Script %s\n", scriptName)

		scriptMap := map[string]interface{}{
			"userId":     hc.ActionSummary.UserId,
			"action":     hc.ActionSummary.Action,
			"collection": hc.ActionSummary.Collection,
			"id":         hc.ActionSummary.Pk,
			"fields":     hc.ActionSummary.Fields,
		}

		_, err := hc.Core.RunScript(scriptName, scriptMap, hc.DB)
		if err != nil {
			log.Println(err.Error())
			return
		}
		log.Println("Hook script complete")

	}
	if hook.Email != nil {

		log.Println("Send Email " + hook.Email.Template + " TO " + hook.Email.Recipient)

		report, err := hc.Core.GetReportHTMLWriter(hook.Email.Template, hc.ActionSummary.Pk, hc.Session)
		if err != nil {
			log.Println(err.Error())
			return
		}

		go hc.Core.SendMailFromResponse(report, hook.Email.Recipient, "")

	}
}
