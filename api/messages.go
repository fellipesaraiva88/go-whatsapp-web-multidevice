package handler

import (
	"net/http"
)

func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	msgType := r.URL.Query().Get("type")

	switch msgType {
	case "text":
		SendText(w, r)
	case "image":
		SendImage(w, r)
	case "audio":
		SendAudio(w, r)
	case "file":
		SendFile(w, r)
	case "contact":
		SendContact(w, r)
	case "location":
		SendLocation(w, r)
	case "poll":
		SendPoll(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Message type not found"))
	}
}

