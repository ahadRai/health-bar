package handlers

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
)

type ServiceConfig struct {
    AuthServiceURL        string
    PatientServiceURL     string
    DoctorServiceURL      string
    TimelineServiceURL    string
    PrescriptionServiceURL string
}

type ProxyHandler struct {
    config ServiceConfig
}

func NewProxyHandler(config ServiceConfig) *ProxyHandler {
    return &ProxyHandler{config: config}
}

// ProxyRequest forwards requests to the appropriate microservice
func (h *ProxyHandler) ProxyRequest(w http.ResponseWriter, r *http.Request) {
    // Determine target service based on URL path
    targetURL := h.getTargetURL(r.URL.Path)
    if targetURL == "" {
        http.Error(w, `{"success":false,"error":"Service not found"}`, http.StatusNotFound)
        return
    }

    // Build full target URL
    fullURL := targetURL + r.URL.Path + "?" + r.URL.RawQuery

    // Create new request
    proxyReq, err := http.NewRequest(r.Method, fullURL, r.Body)
    if err != nil {
        log.Printf("Error creating proxy request: %v", err)
        http.Error(w, `{"success":false,"error":"Internal server error"}`, http.StatusInternalServerError)
        return
    }

    // Copy headers from original request
    for key, values := range r.Header {
        for _, value := range values {
            proxyReq.Header.Add(key, value)
        }
    }

    // Add X-Forwarded-For header
    proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)

    // Send request to target service
    client := &http.Client{}
    resp, err := client.Do(proxyReq)
    if err != nil {
        log.Printf("Error forwarding request to %s: %v", fullURL, err)
        http.Error(w, `{"success":false,"error":"Service unavailable"}`, http.StatusServiceUnavailable)
        return
    }
    defer resp.Body.Close()

    // Copy response headers
    for key, values := range resp.Header {
        for _, value := range values {
            w.Header().Add(key, value)
        }
    }

    // Copy status code
    w.WriteHeader(resp.StatusCode)

    // Copy response body
    io.Copy(w, resp.Body)
}

// getTargetURL determines which service to route to based on the path
func (h *ProxyHandler) getTargetURL(path string) string {
    switch {
    case strings.HasPrefix(path, "/api/auth"):
        return h.config.AuthServiceURL
    case strings.HasPrefix(path, "/api/patients"):
        return h.config.PatientServiceURL
    case strings.HasPrefix(path, "/api/doctors"):
        return h.config.DoctorServiceURL
    case strings.HasPrefix(path, "/api/timeline"):
        return h.config.TimelineServiceURL
    case strings.HasPrefix(path, "/api/prescriptions"):
        return h.config.PrescriptionServiceURL
    default:
        return ""
    }
}

// HealthCheck returns gateway health status
func (h *ProxyHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Check all services
    services := map[string]string{
        "auth":         h.config.AuthServiceURL + "/api/auth/login",
        "patient":      h.config.PatientServiceURL + "/api/patients/profile",
        "doctor":       h.config.DoctorServiceURL + "/api/doctors/profile",
        "timeline":     h.config.TimelineServiceURL + "/api/timeline/my",
        "prescription": h.config.PrescriptionServiceURL + "/api/prescriptions/my",
    }

    status := make(map[string]string)
    allHealthy := true

    for name, url := range services {
        resp, err := http.Get(url)
        if err != nil || resp.StatusCode >= 500 {
            status[name] = "unhealthy"
            allHealthy = false
        } else {
            status[name] = "healthy"
        }
        if resp != nil {
            resp.Body.Close()
        }
    }

    statusCode := http.StatusOK
    if !allHealthy {
        statusCode = http.StatusServiceUnavailable
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    fmt.Fprintf(w, `{"success":%t,"services":%v}`, allHealthy, formatJSON(status))
}

// formatJSON is a simple helper to format map as JSON
func formatJSON(m map[string]string) string {
    var pairs []string
    for k, v := range m {
        pairs = append(pairs, fmt.Sprintf(`"%s":"%s"`, k, v))
    }
    return "{" + strings.Join(pairs, ",") + "}"
}
