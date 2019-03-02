package gddns

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"appengine"
	"appengine/urlfetch"
	"appengine/user"
)

var (
	templates = template.Must(template.ParseFiles(
		"domain_list.html",
	))
)

func init() {
	http.HandleFunc("/admin/domains/list", listDomains)
	http.HandleFunc("/admin/domains/add", addDomains)
	http.HandleFunc("/update_ip", updateIp)
}

func listDomains(w http.ResponseWriter, r *http.Request) {
	u := getAdminUser(w, r)
	if u == nil {
		return
	}

	c := appengine.NewContext(r)
	domains, err := ListDomains(c, u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	b := &bytes.Buffer{}
	if err := templates.ExecuteTemplate(b, "domain_list.html", domains); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	b.WriteTo(w)
}

func addDomains(w http.ResponseWriter, r *http.Request) {
	u := getAdminUser(w, r)
	if u == nil {
		return
	}

	c := appengine.NewContext(r)
	if _, err := AddDomain(c, u, r.FormValue("hostname"), r.FormValue("username"), r.FormValue("password")); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/admin/domains/list", http.StatusFound)
}

func updateIp(w http.ResponseWriter, r *http.Request) {
	domainKey := r.FormValue("domain_key")
	if domainKey == "" {
		http.Error(w, "domain_key parameter must be specified", http.StatusInternalServerError)
		return
	}

	c := appengine.NewContext(r)
	domain, err := GetDomain(c, domainKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if domain == nil {
		http.Error(w, "domain_key not found", http.StatusNotFound)
		return
	}
	c.Infof("Found DomainConfig: %v", domain)

	myip, err := ipFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Infof("Update IP for %v to %v", domainKey, myip)

	// Build Google Domains URL used to update the IP
	var ipUpdateUrl *url.URL
	ipUpdateUrl, err = url.Parse("https://domains.google.com")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ipUpdateUrl.Path += "/nic/update"
	parameters := url.Values{}
	parameters.Add("hostname", domain.Hostname)
	parameters.Add("myip", myip)
	ipUpdateUrl.RawQuery = parameters.Encode()

	c.Infof("Generated Dynamic DNS Update URL: %v", ipUpdateUrl)

	// Build urlfetch client to make the Google Domains request
	client := urlfetch.Client(c)
	req, err := http.NewRequest("GET", ipUpdateUrl.String(), nil)
	req.SetBasicAuth(domain.Username, domain.Password)
	resp, err := client.Do(req)
	if err != nil {
		c.Errorf("Failed to update DNS entry: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Errorf("Failed to read DNS update response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Copy status code and body from Google DNS
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func getAdminUser(w http.ResponseWriter, r *http.Request) *user.User {
	u := getCurrentUser(w, r)
	if u == nil {
		return nil
	}
	if !u.Admin {
		http.Error(w, "Admin Access Only", http.StatusUnauthorized)
		return nil
	}
	return u
}

func getCurrentUser(w http.ResponseWriter, r *http.Request) *user.User {
	c := appengine.NewContext(r)
	u := user.Current(c)

	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return nil
	}

	return u
}

// Get the IP to update the domain to
func ipFromRequest(r *http.Request) (string, error) {
	myip := r.FormValue("myip")
	if myip != "" {
		return myip, nil
	}

	userIP := net.ParseIP(r.RemoteAddr)
	if userIP == nil {
		return "", fmt.Errorf("userip: %q is not a valid IP", r.RemoteAddr)
	}
	return userIP.String(), nil
}
