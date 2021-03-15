package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
)

type Interviewer struct {
	Name string `json:"name"`
	Date string `json:"date"`
	Hour string `json:"hour"`
}

func (i Interviewer) String() string {

	return fmt.Sprintf("%v\n%v\n%v\n", i.Name, i.Date, i.Hour)
}

func (i Interviewer) UnmarshalBinary(data []byte) error {

	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}

	return nil
}

func (i Interviewer) MarshalBinary() (data []byte, err error) {

	return json.Marshal(&i)
}

type Candidate struct {
	Name string `json:"name"`
	Date string `json:"date"`
	Hour string `json:"hour"`
}

func (c Candidate) String() string {
	return fmt.Sprintf("%v\n%v\n%v\n", c.Name, c.Date, c.Hour)
}

func (c Candidate) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &c); err != nil {
		return err
	}

	return nil
}

func (c Candidate) MarshalBinary() (data []byte, err error) {
	return json.Marshal(&c)
}

var client *redis.Client
var tmp *template.Template

func initRedis() *redis.Client {
	client = redis.NewClient(&redis.Options{
		Addr: "db:6379",
	})
	client.Del("Candidate", "Interviewer", "Interviews")
	return client
}

func CandidateGet(rw http.ResponseWriter, r *http.Request) {

	tmp.ExecuteTemplate(rw, "candidate.html", nil)

}

func CandidatePost(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	c := r.Form.Get("candidate")
	day := r.Form.Get("Day")
	appt := r.Form.Get("Hours")

	can := Candidate{
		Name: c,
		Date: day,
		Hour: appt,
	}

	if c == "" {
		http.Redirect(rw, r, "/candidate/"+r.Form.Get("candidateForm2"), http.StatusFound)
		return
	}
	err := client.LPush("Candidate", can).Err()
	if err != nil {
		fmt.Println(err)

	}

	http.Redirect(rw, r, "/candidate/"+c, http.StatusFound)
}

func CandidateGetSlots(rw http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	allCandidates := client.LRange("Candidate", 0, 30).Val()

	allCandidatesUnmarshal := make([]Candidate, len(allCandidates))
	for i, candidate := range allCandidates {
		err := json.Unmarshal([]byte(candidate), &allCandidatesUnmarshal[i])
		if err != nil {
			fmt.Println(err)
		}
	}

	singleCandidate := make([]Candidate, 0)

	for _, candidate := range allCandidatesUnmarshal {
		if candidate.Name == name {
			singleCandidate = append(singleCandidate, candidate)
		}
	}
	//Search Interviewer
	allInterviewers := client.LRange("Interviewer", 0, 30).Val()

	allInterviewersUnmarshal := make([]Interviewer, len(allInterviewers))

	for i, interviewer := range allInterviewers {
		err := json.Unmarshal([]byte(interviewer), &allInterviewersUnmarshal[i])
		if err != nil {
			fmt.Println(err)
		}
	}

	interviewerWithSameTime := make([]Interviewer, 0)

	for _, interviewer := range allInterviewersUnmarshal {
		for _, candidate := range singleCandidate {
			if interviewer.Date == candidate.Date && interviewer.Hour == candidate.Hour {
				interviewerWithSameTime = append(interviewerWithSameTime, interviewer)
			}
		}
	}

	args := struct {
		Candidate    []Candidate
		Interviewers []Interviewer
	}{
		singleCandidate,
		interviewerWithSameTime,
	}

	err := tmp.ExecuteTemplate(rw, "singleCandidate.html", args)

	if err != nil {
		fmt.Println(err)
	}

}

func InterviewerGet(rw http.ResponseWriter, r *http.Request) {

	tmp.ExecuteTemplate(rw, "interviewer.html", nil)
}

func InterviewerPost(rw http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	i := r.Form.Get("interviewer")
	day := r.Form.Get("Day")
	appt := r.Form.Get("Hours")

	can := Interviewer{
		Name: i,
		Date: day,
		Hour: appt,
	}
	if i == "" {
		name := r.Form.Get("interviewerForm2")
		http.Redirect(rw, r, "/interviewer/"+name, http.StatusFound)
		return

	}

	err := client.LPush("Interviewer", can).Err()
	if err != nil {

		fmt.Println(err)

	}

	http.Redirect(rw, r, "/interviewer/"+i, http.StatusFound)

}

func InterviewerGetSlots(rw http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	sl := client.LRange("Interviewer", 0, 30).Val()

	sl2 := make([]Interviewer, len(sl))

	for i, interviewer := range sl {
		err := json.Unmarshal([]byte(interviewer), &sl2[i])
		if err != nil {
			panic(err)
		}
	}

	sl3 := make([]Interviewer, 0)

	for _, interviewer := range sl2 {
		if interviewer.Name == name {
			sl3 = append(sl3, interviewer)
		}
	}

	tmp.ExecuteTemplate(rw, "singleInterviewer.html", sl3)
}

func HomeGet(rw http.ResponseWriter, r *http.Request) {
	sl, err := client.LRange("Interviews", 0, 30).Result()
	if err != nil {
		fmt.Println(err)
	}
	tmp.ExecuteTemplate(rw, "home.html", sl)

}

func InterviewSchedule(rw http.ResponseWriter, r *http.Request) {
	candidateName := mux.Vars(r)["name"]
	r.ParseForm()
	formValue := r.Form.Get("interviewer")

	interviewerAsStr := strings.Split(formValue, "\n")

	if len(interviewerAsStr) != 4 {
		return
	}
	interviewer := Interviewer{
		Name: interviewerAsStr[0][0 : len(interviewerAsStr[0])-1],
		Date: interviewerAsStr[1][0 : len(interviewerAsStr[1])-1],
		Hour: interviewerAsStr[2][0 : len(interviewerAsStr[2])-1],
	}

	e, _ := client.LRem("Interviewer", 1, interviewer).Result()

	fmt.Println(e)
	client.LPush("Interviews", "Candidate "+candidateName+" with "+"the interviewer "+formValue)

	candidateSlots, err := client.LRange("Candidate", 0, 30).Result()

	if err != nil {
		fmt.Println(err)
	}
	candidateSlotsUnmarshal := make([]Candidate, len(candidateSlots))

	for _, candidate := range candidateSlots {
		c := Candidate{}
		err := json.Unmarshal([]byte(candidate), &c)
		if err != nil {
			fmt.Println(err)
			continue
		}
		candidateSlotsUnmarshal = append(candidateSlotsUnmarshal, c)
	}

	for _, candidate := range candidateSlotsUnmarshal {
		if candidate.Name == candidateName {
			client.LRem("Candidate", 1, candidate)
		}
	}

	http.Redirect(rw, r, "/", http.StatusFound)

}

func myNewRouter() *mux.Router {

	r := mux.NewRouter()
	r.HandleFunc("/candidate", CandidateGet).Methods("GET")
	r.HandleFunc("/candidate", CandidatePost).Methods("POST")
	r.HandleFunc("/candidate/{name}", CandidateGetSlots).Methods("Get")
	r.HandleFunc("/candidate/{name}", InterviewSchedule).Methods("Post")
	r.HandleFunc("/interviewer", InterviewerPost).Methods("POST")
	r.HandleFunc("/interviewer", InterviewerGet).Methods("GET")
	r.HandleFunc("/interviewer/{name}", InterviewerGetSlots).Methods("Get")
	r.HandleFunc("/", HomeGet).Methods("GET")
	return r
}
func main() {
	client = initRedis()
	tmp = template.Must(template.ParseGlob("templates/*.html"))
	r := myNewRouter()
	fs := http.FileServer(http.Dir("./css/"))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", fs))
	http.Handle("/", r)
	host := ":8080"
	fmt.Println("Serving on localhost" + host)
	fmt.Println(http.ListenAndServe(host, nil))
}
