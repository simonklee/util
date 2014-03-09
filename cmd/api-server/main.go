package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/gorilla/mux"
	"github.com/nu7hatch/gouuid"
	"github.com/simonz05/util/kvstore"
	"github.com/simonz05/util/log"
	"github.com/simonz05/util/session"
)

var (
	help           = flag.Bool("h", false, "show help text")
	configFilename = flag.String("config", "config.toml", "config file path")
	laddr          = flag.String("laddr ", ":8080", "laddr for server")
)

var indexTmpl = template.Must(template.New("index").ParseFiles("templates/base.html", "templates/index.html"))
var keyTmpl = template.Must(template.New("key").ParseFiles("templates/base.html", "templates/key.html"))

type Key struct {
	ID      string
	Created time.Time
}

type Config struct {
	Laddr   string
	Regions map[string]*Region
}

type Region struct {
	Codename, Name string
	Selected       bool
	RedisDSN       string
	Backend        *RedisBackend
}

var (
	config Config
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nOptions:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Description:
  Runs a webserver which manages API keys.

  `)
}

func currentRegion(r *http.Request) *Region {
	cookie, err := r.Cookie("region")

	if err != nil {
		log.Println(err)
	} else {
		codename := cookie.Value

		for k, v := range config.Regions {
			if codename == k {
				b := *v
				return &b
			}
		}
	}

	for _, v := range config.Regions {
		b := *v
		return &b
	}

	panic("no regions")
}

func selectRegionHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	regionString := r.Form.Get("region")
	var region *Region

	fmt.Println("Select region", regionString)

	for k, v := range config.Regions {
		if regionString == k {
			region = v
			break
		}
	}

	if region == nil {
		region = currentRegion(r)
	}

	expire := time.Now().AddDate(0, 0, 30)
	cookie := http.Cookie{
		Name:     "region",
		Value:    region.Codename,
		Path:     "/",
		HttpOnly: true,
		Expires:  expire.UTC(),
	}

	http.SetCookie(w, &cookie)
}

func getRegions(r *http.Request) map[string]*Region {
	dst := make(map[string]*Region, len(config.Regions))
	region := currentRegion(r)

	for k, v := range config.Regions {
		b := *v
		dst[k] = &b
		b.Selected = region.Codename == k
	}

	return dst
}

func getKeys(region *Region) ([]*Key, error) {
	values, err := region.Backend.Get()

	if err != nil {
		log.Errorln(err)
		return nil, err
	}

	keys := make([]*Key, 0, len(values))

	for _, v := range values {
		keys = append(keys, &Key{ID: v})
	}

	return keys, nil
}

func createKey(region *Region) (*Key, error) {
	id, _ := uuid.NewV4()
	key := &Key{ID: fmt.Sprintf("%s", id)}
	perm := &session.Session{}
	perm.Set(session.FullMask)
	err := region.Backend.Set(key.ID, perm)
	return key, err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	region := currentRegion(r)
	keys, err := getKeys(region)

	if err != nil {
		http.Error(w, "Server Error", 501)
		return
	}

	data := struct {
		Keys    []*Key
		Regions map[string]*Region
	}{
		keys, getRegions(r),
	}

	indexTmpl.ExecuteTemplate(w, "base", data)
}

func creatorHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	region := currentRegion(r)
	key, err := createKey(region)

	if err != nil {
		http.Error(w, "Server Error", 501)
		return
	}

	data := struct {
		Key *Key
	}{
		key,
	}

	// redirect to single key page
	keyTmpl.ExecuteTemplate(w, "base", data)
}

func deleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	region := currentRegion(r)
	region.Backend.Delete(vars["key"])
	fmt.Println("DELETE")
}

func runServer() {
	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler).Methods("GET").Name("index")
	router.HandleFunc("/", creatorHandler).Methods("POST").Name("key-create")
	router.HandleFunc("/{key:[0-9A-Za-z-]+}/", deleteHandler).Methods("DELETE").Name("key-delete")
	router.HandleFunc("/select-region/", selectRegionHandler).Methods("POST").Name("select-region")
	log.Printf("Listen on %s\n", config.Laddr)
	http.ListenAndServe(config.Laddr, router)
}

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(1)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	if _, err := toml.DecodeFile(*configFilename, &config); err != nil {
		log.Fatal(err)
	}

	if config.Laddr == "" {
		config.Laddr = *laddr
	}

	for k, v := range config.Regions {
		v.Codename = k
		store, err := kvstore.Open(v.RedisDSN)

		if err != nil {
			log.Fatal(err)
		}

		conn := store.Get()
		_, err = conn.Do("Ping")
		conn.Close()

		if err != nil {
			log.Fatalf("region %s: %v", k, err)
		}

		v.Backend = NewRedisBackend(store, v.Codename)
	}

	runServer()
}
