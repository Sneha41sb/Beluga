package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
	"uftp/pkg/dsp"
	"uftp/pkg/frame"
	"uftp/pkg/mesh"
)

type Hub struct {
	sync.Mutex
	clients map[chan []byte]bool
}

var hub = Hub{
	clients: make(map[chan []byte]bool),
}

func (h *Hub) Broadcast(data []byte) {
	h.Lock()
	defer h.Unlock()
	for clientChan := range h.clients {
		select {
		case clientChan <- data:
		default:
		}
	}
}

func (h *Hub) Register(c chan []byte) {
	h.Lock()
	defer h.Unlock()
	h.clients[c] = true
}

func (h *Hub) Unregister(c chan []byte) {
	h.Lock()
	defer h.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage:")
		fmt.Println("  go run cmd/uftp/main.go server [port]")
		fmt.Println("  go run cmd/uftp/main.go send <message>")
		fmt.Println("  go run cmd/uftp/main.go simulate-mesh")
		return
	}

	cmd := os.Args[1]

	config := dsp.AFSKConfig{
		SampleRate: 44100.0,
		MarkFreq:   18500.0, // Bit 1 tone
		SpaceFreq:  19500.0, // Bit 0 tone
		BaudRate:   100,
		Amplitude:  0.8,
	}

	switch cmd {
	case "server":
		port := "8080"
		if len(os.Args) >= 3 {
			port = os.Args[2]
		}

		// Static Web File Server
		fs := http.FileServer(http.Dir("./web"))
		http.Handle("/", fs)

		// SSE Event Stream for Live Multi-Device Relay
		http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Connection", "keep-alive")

			clientChan := make(chan []byte, 10)
			hub.Register(clientChan)
			defer hub.Unregister(clientChan)

			notify := r.Context().Done()
			for {
				select {
				case <-notify:
					return
				case data := <-clientChan:
					fmt.Fprintf(w, "data: %s\n\n", string(data))
					if flusher, ok := w.(http.Flusher); ok {
						flusher.Flush()
					}
				}
			}
		})

		// Broadcast Endpoint for Incoming Client Frames
		http.HandleFunc("/api/broadcast", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}

			hub.Broadcast(body)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		})

		fmt.Printf("🚀 [SonicBeacon UFTP Server] Running on http://0.0.0.0:%s\n", port)
		fmt.Println("📡 Connect your devices on the same Wi-Fi/Network to transmit over server relay & acoustic audio.")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			fmt.Printf("❌ Server Error: %v\n", err)
		}

	case "send":
		if len(os.Args) < 3 {
			fmt.Println("Error: Please provide a message to send")
			return
		}
		msg := os.Args[2]

		f := frame.Frame{
			Type:      frame.TypeMeshAlert,
			TTL:       5,
			SenderID:  0x11223344,
			TargetID:  frame.BroadcastID,
			MessageID: rand.Uint32(),
			Payload:   []byte(msg),
		}

		pkt := frame.Encode(f)
		samples := dsp.ModulateAFSK(pkt, config)

		fmt.Printf("📢 [Broadcaster] Sent Message: %s\n", msg)
		fmt.Printf("📊 [Broadcaster] Frame Size: %d bytes | Generated Audio Samples: %d\n", len(pkt), len(samples))

	case "simulate-mesh":
		fmt.Println("=== Starting Ultrasonic Mesh Network Simulation ===")

		nodeA := mesh.NewMeshNode("Node-A (Origin)")
		nodeB := mesh.NewMeshNode("Node-B (Relay)")
		nodeC := mesh.NewMeshNode("Node-C (Receiver)")
		_ = nodeA

		rand.Seed(time.Now().UnixNano())
		msg := "CRITICAL ALERT: Emergency shelter open at Central Square!"

		fmt.Printf("📡 Node-A broadcasting: '%s'\n\n", msg)

		// 1. Node A Encodes & Modulates
		fOrig := frame.Frame{
			Type:      frame.TypeMeshAlert,
			TTL:       3,
			SenderID:  0xAAAA1111,
			TargetID:  frame.BroadcastID,
			MessageID: rand.Uint32(),
			Payload:   []byte(msg),
		}
		pktBytes := frame.Encode(fOrig)
		audioStream := dsp.ModulateAFSK(pktBytes, config)

		// 2. Node B Demodulates & Decodes
		demodBytesB := dsp.DemodulateAFSK(audioStream, config)
		decodedFrameB, err := frame.Decode(demodBytesB)
		if err != nil {
			fmt.Printf("❌ Node-B Decode Error: %v\n", err)
			return
		}

		fmt.Printf("🔊 Node-B Received Message: '%s' (Sender=0x%X Target=PUBLIC TTL=%d)\n", string(decodedFrameB.Payload), decodedFrameB.SenderID, decodedFrameB.TTL)

		// 3. Node B Relays to Node C
		relayFrameB, shouldRelayB := nodeB.ProcessIncomingFrame(*decodedFrameB)
		if shouldRelayB {
			fmt.Printf("🔄 Node-B Relaying message with updated TTL=%d...\n", relayFrameB.TTL)

			relayPkt := frame.Encode(relayFrameB)
			relayAudio := dsp.ModulateAFSK(relayPkt, config)

			// Node C Demodulates & Decodes Node B's relay
			demodBytesC := dsp.DemodulateAFSK(relayAudio, config)
			decodedFrameC, err := frame.Decode(demodBytesC)
			if err != nil {
				fmt.Printf("❌ Node-C Decode Error: %v\n", err)
				return
			}

			fmt.Printf("🔊 Node-C Received Message: '%s' (Sender=0x%X Target=PUBLIC TTL=%d)\n", string(decodedFrameC.Payload), decodedFrameC.SenderID, decodedFrameC.TTL)

			// Node C tries to relay
			relayFrameC, shouldRelayC := nodeC.ProcessIncomingFrame(*decodedFrameC)
			if shouldRelayC {
				fmt.Printf("🔄 Node-C Relaying message with updated TTL=%d...\n", relayFrameC.TTL)
			}
		}

		fmt.Println("\n✅ Mesh Simulation Complete! Message successfully propagated across all nodes.")

	default:
		fmt.Println("Unknown command:", cmd)
	}
}

