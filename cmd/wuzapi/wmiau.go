package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/mdp/qrterminal/v3"
	"github.com/patrickmn/go-cache"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	_ "modernc.org/sqlite"

	//	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	//"google.golang.org/protobuf/proto"
)

// var wlog waLog.Logger
var clientPointer = make(map[int]*whatsmeow.Client)
var clientHttp = make(map[int]*resty.Client)
var historySyncID int32

type MyClient struct {
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
	userID         int
	token          string
	subscriptions  []string
	db             *sql.DB
}

// Connects to Whatsapp Websocket on server startup if last state was connected
func (s *server) connectOnStartup() {
	rows, err := s.db.Query("SELECT id, token, jid, webhook, events, osname, platformtype FROM users WHERE connected=1")
	if err != nil {
		log.Error().Err(err).Msg("DB Problem")
		return
	}
	defer rows.Close()

	for rows.Next() {
		txtid := ""
		token := ""
		jid := ""
		webhook := ""
		events := ""
		osName := ""
		platformType := ""

		err = rows.Scan(&txtid, &token, &jid, &webhook, &events, &osName, &platformType)
		if err != nil {
			log.Error().Err(err).Msg("DB Problem")
			return
		} else {
			log.Info().Str("token", token).Msg("Connect to Whatsapp on startup")
			v := Values{map[string]string{
				"Id":           txtid,
				"Jid":          jid,
				"Webhook":      webhook,
				"Token":        token,
				"Events":       events,
				"OSName":       osName,
				"PlatformType": platformType,
			}}
			userinfocache.Set(token, v, cache.NoExpiration)

			userid, _ := strconv.Atoi(txtid)

			// Gets and set subscription to webhook events
			eventarray := strings.Split(events, ",")
			var subscribedEvents []string
			if len(eventarray) < 1 {
				if !Find(subscribedEvents, "All") {
					subscribedEvents = append(subscribedEvents, "All")
				}
			} else {
				for _, arg := range eventarray {
					if !Find(messageTypes, arg) {
						log.Warn().Str("Type", arg).Msg("Message type discarded")
						continue
					}
					if !Find(subscribedEvents, arg) {
						subscribedEvents = append(subscribedEvents, arg)
					}
				}
			}

			eventstring := strings.Join(subscribedEvents, ",")
			log.Info().Str("events", eventstring).Str("jid", jid).Msg("Attempt to connect")
			killchannel[userid] = make(chan bool)
			go s.startClient(userid, jid, token, subscribedEvents, osName, platformType)
		}
	}

	err = rows.Err()
	if err != nil {
		log.Error().Err(err).Msg("DB Problem")
	}
}

func parseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			log.Error().Err(err).Msg("Invalid JID")
			return recipient, false
		} else if recipient.User == "" {
			log.Error().Err(err).Msg("Invalid JID no server specified")
			return recipient, false
		}
		return recipient, true
	}
}

func (s *server) startClient(userID int, textjid string, token string, subscriptions []string, osName string, platformType string) {

	log.Info().Str("userid", strconv.Itoa(userID)).Str("jid", textjid).Msg("Starting websocket connection to Whatsapp")

	var deviceStore *store.Device
	var err error

	if clientPointer[userID] != nil {
		isConnected := clientPointer[userID].IsConnected()
		if isConnected {
			return
		}
	}

	if textjid != "" {
		jid, _ := parseJID(textjid)
		// If you want multiple sessions, remember their JIDs and use .GetDevice(jid) or .GetAllDevices() instead.
		//deviceStore, err := container.GetFirstDevice()
		deviceStore, err = container.GetDevice(jid)
		if err != nil {
			panic(err)
		}
	} else {
		log.Warn().Msg("No jid found. Creating new device")
		deviceStore = container.NewDevice()
	}

	if deviceStore == nil {
		log.Warn().Msg("No store found. Creating new one")
		deviceStore = container.NewDevice()
	}

	//store.CompanionProps.PlatformType = waProto.CompanionProps_CHROME.Enum()
	//store.CompanionProps.Os = proto.String("Mac OS")
	// Default values
	defaultOSName := "Mac OS 10"
	defaultPlatformType := "CHROME"

	// Use the provided values or default values if not provided
	if osName == "" {
		osName = defaultOSName
	}

	if platformType == "" {
		platformType = defaultPlatformType
	}

	// Define a map to map platform type strings to enum values
	var platformTypeMap = map[string]waProto.DeviceProps_PlatformType{
		"UNKNOWN":           waProto.DeviceProps_UNKNOWN,
		"CHROME":            waProto.DeviceProps_CHROME,
		"FIREFOX":           waProto.DeviceProps_FIREFOX,
		"IE":                waProto.DeviceProps_IE,
		"OPERA":             waProto.DeviceProps_OPERA,
		"SAFARI":            waProto.DeviceProps_SAFARI,
		"EDGE":              waProto.DeviceProps_EDGE,
		"DESKTOP":           waProto.DeviceProps_DESKTOP,
		"IPAD":              waProto.DeviceProps_IPAD,
		"ANDROID_TABLET":    waProto.DeviceProps_ANDROID_TABLET,
		"OHANA":             waProto.DeviceProps_OHANA,
		"ALOHA":             waProto.DeviceProps_ALOHA,
		"CATALINA":          waProto.DeviceProps_CATALINA,
		"TCL_TV":            waProto.DeviceProps_TCL_TV,
		"IOS_PHONE":         waProto.DeviceProps_IOS_PHONE,
		"IOS_CATALYST":      waProto.DeviceProps_IOS_CATALYST,
		"ANDROID_PHONE":     waProto.DeviceProps_ANDROID_PHONE,
		"ANDROID_AMBIGUOUS": waProto.DeviceProps_ANDROID_AMBIGUOUS,
		"WEAR_OS":           waProto.DeviceProps_WEAR_OS,
		"AR_WRIST":          waProto.DeviceProps_AR_WRIST,
		"AR_DEVICE":         waProto.DeviceProps_AR_DEVICE,
		"UWP":               waProto.DeviceProps_UWP,
		"VR":                waProto.DeviceProps_VR,
	}

	// Convert platformType to uppercase
	platformType = strings.ToUpper(platformType)
	// Retrieve the corresponding enum value from the map
	enumValue, exists := platformTypeMap[platformType]
	if !exists {
		// Handle the case when an invalid platform type is supplied
		enumValue = waProto.DeviceProps_UNKNOWN
	}
	// Set the PlatformType field of DeviceProps to the enum value
	store.DeviceProps.PlatformType = enumValue.Enum()

	store.DeviceProps.Os = &osName

	clientLog := waLog.Stdout("Client", *waDebug, true)
	var client *whatsmeow.Client
	if *waDebug != "" {
		client = whatsmeow.NewClient(deviceStore, clientLog)
	} else {
		client = whatsmeow.NewClient(deviceStore, nil)
	}
	clientPointer[userID] = client
	mycli := MyClient{client, 1, userID, token, subscriptions, s.db}
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.myEventHandler)

	// Initialize the HTTP client
	clientHttp[userID] = resty.New()
	clientHttp[userID].SetRedirectPolicy(resty.FlexibleRedirectPolicy(15))

	// Enable debug logging if waDebug is set to "DEBUG"
	if *waDebug == "DEBUG" {
		clientHttp[userID].SetDebug(true)
	}

	// Set a timeout for requests
	clientHttp[userID].SetTimeout(5 * time.Second)

	// Configure TLS settings
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13, // Use TLS 1.3 for the highest security
	}
	clientHttp[userID].SetTLSClientConfig(tlsConfig)

	// Set error handling for the HTTP client
	clientHttp[userID].OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Debug().Str("response", v.Response.String()).Msg("resty error")
			log.Error().Err(v.Err).Msg("resty error")
		} else {
			// Log other types of errors
			log.Error().Err(err).Msg("resty error")
		}
	})

	if client.Store.ID == nil {
		// No ID stored, new login

		qrChan, err := client.GetQRChannel(context.Background())
		if err != nil {
			// This error means that we're already logged in, so ignore it.
			if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
				log.Error().Err(err).Msg("Failed to get QR channel")
			}
		} else {
			err = client.Connect() // Si no conectamos no se puede generar QR
			if err != nil {
				panic(err)
			}
			for evt := range qrChan {
				if evt.Event == "code" {
					// Display QR code in terminal (useful for testing/developing)
					if *logType != "json" {
						qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
						fmt.Println("QR code:\n", evt.Code)
					}
					// Store encoded/embedded base64 QR on database for retrieval with the /qr endpoint
					image, _ := qrcode.Encode(evt.Code, qrcode.Medium, 256)
					base64qrcode := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)
					var sqlStmt string

					switch dbType {
					case "sqlite3":
						sqlStmt = `UPDATE users SET qrcode=? WHERE id=?`
					case "postgresql":
						sqlStmt = `UPDATE users SET qrcode=$1 WHERE id=$2`
					default:
						log.Error().Err(err).Msg(
							"Failed to store encoded/embedded base64 QR on database for retrieval with the /qr endpoint. Unsupported database")
						return
					}

					_, err := s.db.Exec(sqlStmt, base64qrcode, userID)
					if err != nil {
						log.Error().Err(err).Msg(sqlStmt)
					}

				} else if evt.Event == "timeout" {
					var sqlStmt string

					// Determine the SQL statement based on the database type
					switch dbType {
					case "sqlite3":
						sqlStmt = `UPDATE users SET qrcode=? WHERE id=?`
					case "postgresql":
						sqlStmt = `UPDATE users SET qrcode=$1 WHERE id=$2`
					default:
						log.Error().Msg("Unsupported database type for clearing QR code")
						return
					}

					// Execute the SQL statement to clear the QR code
					_, err := s.db.Exec(sqlStmt, "", userID)
					if err != nil {
						log.Error().Err(err).Msg("Error executing SQL statement to clear QR code")
					}

					// Additional logic for handling timeout
					log.Warn().Msg("QR timeout killing channel")
					delete(clientPointer, userID)
					killchannel[userID] <- true
				} else if evt.Event == "success" {
					log.Info().Msg("QR pairing ok!")

					var sqlStmt string

					// Determine the SQL statement based on the database type
					switch dbType {
					case "sqlite3":
						sqlStmt = `UPDATE users SET qrcode=? WHERE id=?`
					case "postgresql":
						sqlStmt = `UPDATE users SET qrcode=$1 WHERE id=$2`
					default:
						log.Error().Msg("Unsupported database type for updating QR code")
						return
					}

					// Execute the SQL statement to clear the QR code
					_, err := s.db.Exec(sqlStmt, "", userID)
					if err != nil {
						log.Error().Err(err).Msg("Error executing SQL statement to clear QR code")
					}

				} else {
					log.Info().Str("event", evt.Event).Msg("Login event")
				}
			}
		}

	} else {
		// Already logged in, just connect
		log.Info().Msg("Already logged in, just connect")
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// Keep connected client live until disconnected/killed
	for {
		select {
		case <-killchannel[userID]:
			log.Info().Str("userid", strconv.Itoa(userID)).Msg("Received kill signal")

			// Disconnect the client
			client.Disconnect()

			// Remove the client from the pointer map
			delete(clientPointer, userID)

			// Determine the SQL statement based on the database type
			var sqlStmt string
			switch dbType {
			case "sqlite3":
				sqlStmt = `UPDATE users SET connected=0 WHERE id=?`
			case "postgresql":
				sqlStmt = `UPDATE users SET connected=0 WHERE id=$1`
			default:
				log.Error().Msg("Unsupported database type for updating connection status")
				return
			}

			// Execute the SQL statement to update connection status
			_, err := s.db.Exec(sqlStmt, userID)
			if err != nil {
				log.Error().Err(err).Msg("Error executing SQL statement to update connection status")
			}

			return
		default:
			time.Sleep(1000 * time.Millisecond)
			// Uncomment the following line if you want to log the loop
			// log.Info().Str("jid", textjid).Msg("Loop the loop")
		}

	}
}

// Add this helper function
func downloadAndSaveMedia(mycli *MyClient, evt *events.Message, mediaType string, getData func() ([]byte, error), getMimeType func() string, exPath string) (string, error) {
	txtid := strconv.Itoa(mycli.userID)
	userDirectory := filepath.Join(exPath, "files", "user_"+txtid)

	if err := os.MkdirAll(userDirectory, 0751); err != nil {
		return "", fmt.Errorf("could not create user directory: %w", err)
	}

	data, err := getData()
	if err != nil {
		return "", fmt.Errorf("failed to download %s: %w", mediaType, err)
	}

	exts, _ := mime.ExtensionsByType(getMimeType())
	path := filepath.Join(userDirectory, evt.Info.ID+exts[0])

	if err := os.WriteFile(path, data, 0600); err != nil {
		return "", fmt.Errorf("failed to save %s: %w", mediaType, err)
	}

	return path, nil
}

func (mycli *MyClient) myEventHandler(rawEvt interface{}) {
	txtid := strconv.Itoa(mycli.userID)
	postmap := make(map[string]interface{})
	postmap["event"] = rawEvt
	dowebhook := 0
	path := ""

	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)

	switch evt := rawEvt.(type) {
	case *events.AppStateSyncComplete:
		if len(mycli.WAClient.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
			err := mycli.WAClient.SendPresence(types.PresenceAvailable)
			if err != nil {
				log.Warn().Err(err).Msg("Failed to send available presence")
			} else {
				log.Info().Msg("Marked self as available")
			}
		}
	case *events.Connected, *events.PushNameSetting:
		if len(mycli.WAClient.Store.PushName) == 0 {
			return
		}

		// Send presence available when connecting and when the pushname is changed.
		// This ensures that outgoing messages always have the correct pushname.
		err := mycli.WAClient.SendPresence(types.PresenceAvailable)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to send available presence")
		} else {
			log.Info().Msg("Marked self as available")
		}

		// Determine the SQL statement based on the database type
		var sqlStmt string
		switch dbType {
		case "sqlite3":
			sqlStmt = `UPDATE users SET connected=1 WHERE id=?`
		case "postgresql":
			sqlStmt = `UPDATE users SET connected=1 WHERE id=$1`
		default:
			log.Error().Msg("Unsupported database type for updating connection status")
			return
		}

		// Execute the SQL statement to update connection status
		_, err = mycli.db.Exec(sqlStmt, mycli.userID)
		if err != nil {
			log.Error().Err(err).Msg("Error executing SQL statement to update connection status")
			return
		}

	case *events.PairSuccess:
		log.Info().
			Str("userid", strconv.Itoa(mycli.userID)).
			Str("token", mycli.token).
			Str("ID", evt.ID.String()).
			Str("BusinessName", evt.BusinessName).
			Str("Platform", evt.Platform).
			Msg("QR Pair Success")

		jid := evt.ID

		// Determine the SQL statement based on the database type
		var sqlStmt string
		switch dbType {
		case "sqlite3":
			sqlStmt = `UPDATE users SET jid=? WHERE id=?`
		case "postgresql":
			sqlStmt = `UPDATE users SET jid=$1 WHERE id=$2`
		default:
			log.Error().Msg("Unsupported database type for updating JID")
			return
		}

		// Execute the SQL statement to update the JID
		_, err := mycli.db.Exec(sqlStmt, jid, mycli.userID)
		if err != nil {
			log.Error().Err(err).Msg("Error executing SQL statement to update JID")
			return
		}

		myuserinfo, found := userinfocache.Get(mycli.token)
		if !found {
			log.Warn().Msg("No user info cached on pairing?")
		} else {
			txtid := myuserinfo.(Values).Get("Id")
			token := myuserinfo.(Values).Get("Token")
			v := updateUserInfo(myuserinfo, "Jid", jid.String()) // Use String() method for conversion
			userinfocache.Set(token, v, cache.NoExpiration)
			log.Info().Str("jid", jid.String()).Str("userid", txtid).Str("token", token).Msg("User information set")
		}
	case *events.StreamReplaced:
		log.Info().Msg("Received StreamReplaced event")
		return
	case *events.Message:
		postmap["type"] = "Message"
		dowebhook = 1
		metaParts := []string{fmt.Sprintf("pushname: %s", evt.Info.PushName), fmt.Sprintf("timestamp: %s", evt.Info.Timestamp)}
		if evt.Info.Type != "" {
			metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
		}
		if evt.Info.Category != "" {
			metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
		}
		if evt.IsViewOnce {
			metaParts = append(metaParts, "view once")
		}
		if evt.IsViewOnce {
			metaParts = append(metaParts, "ephemeral")
		}

		log.Info().Str("id", evt.Info.ID).Str("source", evt.Info.SourceString()).Str("parts", strings.Join(metaParts, ", ")).Msg("Message Received")

		if img := evt.Message.GetImageMessage(); img != nil {
			path, err := downloadAndSaveMedia(mycli, evt, "image",
				func() ([]byte, error) { return mycli.WAClient.Download(img) },
				img.GetMimetype,
				exPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to handle image")
			} else {
				log.Info().Str("path", path).Msg("Image saved")
			}
		}

		if audio := evt.Message.GetAudioMessage(); audio != nil {
			path, err := downloadAndSaveMedia(mycli, evt, "audio",
				func() ([]byte, error) { return mycli.WAClient.Download(audio) },
				audio.GetMimetype,
				exPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to handle audio")
			} else {
				log.Info().Str("path", path).Msg("Audio saved")
			}
		}

		if document := evt.Message.GetDocumentMessage(); document != nil {
			path, err := downloadAndSaveMedia(mycli, evt, "document",
				func() ([]byte, error) { return mycli.WAClient.Download(document) },
				document.GetMimetype,
				exPath)
			if err != nil {
				log.Error().Err(err).Msg("Failed to handle document")
			} else {
				log.Info().Str("path", path).Msg("Document saved")
			}
		}
	case *events.Receipt:
		postmap["type"] = "ReadReceipt"
		dowebhook = 1
		if evt.Type == types.ReceiptTypeRead || evt.Type == types.ReceiptTypeReadSelf {
			log.Info().Strs("id", evt.MessageIDs).Str("source", evt.SourceString()).Time("timestamp", evt.Timestamp).Msg("Message was read")
			if evt.Type == types.ReceiptTypeRead {
				postmap["state"] = "Read"
			} else {
				postmap["state"] = "ReadSelf"
			}
		} else if evt.Type == types.ReceiptTypeDelivered {
			postmap["state"] = "Delivered"
			log.Info().Str("id", evt.MessageIDs[0]).Str("source", evt.SourceString()).Str("timestamp", fmt.Sprintf("%d", evt.Timestamp.Unix())).Msg("Message delivered")
		} else {
			// Discard webhooks for inactive or other delivery types
			return
		}
	case *events.Presence:
		postmap["type"] = "Presence"
		dowebhook = 1
		if evt.Unavailable {
			postmap["state"] = "offline"
			if evt.LastSeen.IsZero() {
				log.Info().Str("from", evt.From.String()).Msg("User is now offline")
			} else {
				log.Info().Str("from", evt.From.String()).Time("lastSeen", evt.LastSeen).Msg("User is now offline")
			}
		} else {
			postmap["state"] = "online"
			log.Info().Str("from", evt.From.String()).Msg("User is now online")
		}
	case *events.HistorySync:
		postmap["type"] = "HistorySync"
		dowebhook = 1

		// check/creates user directory for files
		userDirectory := filepath.Join(exPath, "files", "user_"+txtid)
		_, err := os.Stat(userDirectory)
		if os.IsNotExist(err) {
			errDir := os.MkdirAll(userDirectory, 0751)
			if errDir != nil {
				log.Error().Err(errDir).Msg("Could not create user directory")
				return
			}
		}

		id := atomic.AddInt32(&historySyncID, 1)
		fileName := filepath.Join(userDirectory, "history-"+strconv.Itoa(int(id))+".json")
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open file to write history sync")
			return
		}
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(evt.Data)
		if err != nil {
			log.Error().Err(err).Msg("Failed to write history sync")
			return
		}
		log.Info().Str("filename", fileName).Msg("Wrote history sync")
		_ = file.Close()
	case *events.AppState:
		log.Info().Str("index", fmt.Sprintf("%+v", evt.Index)).Str("actionValue", fmt.Sprintf("%+v", evt.SyncActionValue)).Msg("App state event received")
	case *events.LoggedOut:
		log.Info().Str("reason", evt.Reason.String()).Msg("Logged out")

		// Send kill signal
		killchannel[mycli.userID] <- true

		// Determine the SQL statement based on the database type
		var sqlStmt string
		switch dbType {
		case "sqlite3":
			sqlStmt = `UPDATE users SET connected=0 WHERE id=?`
		case "postgresql":
			sqlStmt = `UPDATE users SET connected=0 WHERE id=$1`
		default:
			log.Error().Msg("Unsupported database type for updating connection status")
			return
		}

		// Execute the SQL statement to update connection status
		_, err := mycli.db.Exec(sqlStmt, mycli.userID)
		if err != nil {
			log.Error().Err(err).Msg("Error executing SQL statement to update connection status")
			return
		}

	case *events.ChatPresence:
		postmap["type"] = "ChatPresence"
		dowebhook = 1
		log.Info().
			Str("state", fmt.Sprintf("%v", evt.State)).
			Str("media", fmt.Sprintf("%v", evt.Media)).
			Str("chat", evt.MessageSource.Chat.String()).
			Str("sender", evt.MessageSource.Sender.String()).
			Msg("Chat Presence received")
	case *events.CallOffer:
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call offer")
	case *events.CallAccept:
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call accept")
	case *events.CallTerminate:
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call terminate")
	case *events.CallOfferNotice:
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call offer notice")
	case *events.CallRelayLatency:
		log.Info().Str("event", fmt.Sprintf("%+v", evt)).Msg("Got call relay latency")
	default:
		log.Warn().Str("event", fmt.Sprintf("%+v", evt)).Msg("Unhandled event")
	}

	if dowebhook == 1 {
		// call webhook
		webhookurl := ""
		myuserinfo, found := userinfocache.Get(mycli.token)
		if !found {
			log.Warn().Str("token", mycli.token).Msg("Could not call webhook as there is no user for this token")
		} else {
			webhookurl = myuserinfo.(Values).Get("Webhook")
		}

		if !Find(mycli.subscriptions, postmap["type"].(string)) && !Find(mycli.subscriptions, "All") {
			log.Warn().Str("type", postmap["type"].(string)).Msg("Skipping webhook. Not subscribed for this type")
			return
		}

		if webhookurl != "" {
			log.Info().Str("url", webhookurl).Msg("Calling webhook")
			values, _ := json.Marshal(postmap)
			data := map[string]string{
				"jsonData": string(values),
				"token":    mycli.token,
			}

			if path == "" {
				go callHook(webhookurl, data, mycli.userID)
			} else {
				// Create a channel to capture error from the goroutine
				errChan := make(chan error, 1)
				go func() {
					err := callHookFile(webhookurl, data, mycli.userID, path)
					errChan <- err
				}()

				// Optionally handle the error from the channel
				if err := <-errChan; err != nil {
					log.Error().Err(err).Msg("Error calling hook file")
				}
			}
		} else {
			log.Warn().Str("userid", strconv.Itoa(mycli.userID)).Msg("No webhook set for user")
		}
	}

}
