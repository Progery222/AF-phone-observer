package handler

type APIVisibility string

const (
	APIPublic  APIVisibility = "public"
	APIPrivate APIVisibility = "private"
)

type APIRoute struct {
	Visibility  APIVisibility
	Transport   string
	Method      string
	Path        string
	Name        string
	Handler     string
	Description string
}

func PublicAPIRoutes() []APIRoute {
	return []APIRoute{
		{APIPublic, "http", "GET", "/health", "health", "HTTPHandler.health", "Liveness probe."},
		{APIPublic, "http", "GET", "/ready", "ready", "HTTPHandler.ready", "Readiness probe."},
		{APIPublic, "http", "POST", "/screenshot", "observer.screenshot", "HTTPHandler.screenshot", "Capture a screenshot."},
		{APIPublic, "http", "POST", "/dump-ui", "observer.dump_ui", "HTTPHandler.dumpUI", "Dump current UI tree."},
		{APIPublic, "http", "POST", "/find-element", "observer.find_element", "HTTPHandler.findElement", "Find an element by text, resource id, description, or hint."},
		{APIPublic, "http", "POST", "/wait-for-element", "observer.wait_for_element", "HTTPHandler.waitForElement", "Poll until an element appears."},
		{APIPublic, "http", "POST", "/detect-state", "observer.detect_state", "HTTPHandler.detectState", "Classify current screen state."},
		{APIPublic, "http", "GET", "/screen/{serial}", "observer.current_screen", "HTTPHandler.currentScreen", "Capture or return current screen for a serial."},
		{APIPublic, "http", "GET", "/ui/{serial}", "observer.current_ui", "HTTPHandler.currentUI", "Capture or return current UI tree for a serial."},
		{APIPublic, "http", "DELETE", "/cache/{serial}", "observer.clear_cache", "HTTPHandler.clearCache", "Clear in-memory observer cache."},
		{APIPublic, "grpc", "rpc", "/observer.v1.ObserverService/CaptureScreenshot", "observer.grpc.capture_screenshot", "ObserverHandler.CaptureScreenshot", "Capture screenshot through gRPC contract."},
		{APIPublic, "grpc", "rpc", "/observer.v1.ObserverService/DumpUI", "observer.grpc.dump_ui", "ObserverHandler.DumpUI", "Dump UI through gRPC contract."},
		{APIPublic, "grpc", "rpc", "/observer.v1.ObserverService/DetectState", "observer.grpc.detect_state", "ObserverHandler.DetectState", "Detect state through gRPC contract."},
	}
}

func PrivateAPIRoutes() []APIRoute {
	return []APIRoute{
		{APIPrivate, "adb", "EXEC", "screencap, uiautomator dump", "adb.observe", "ADB driver", "Read screen pixels and UI XML."},
		{APIPrivate, "s3", "PUT", "MinIO af-screenshots", "storage.screenshots", "storage adapter", "Store screenshots and return object keys."},
		{APIPrivate, "http-out", "POST", "ollama/openai/vision-server", "vlm.detect", "VLM adapter", "Classify screenshots when UI-only detection is not enough."},
		{APIPrivate, "memory", "READ/WRITE", "screen cache", "cache.screen", "cache", "Cache recent observations per serial."},
	}
}

func AllAPIRoutes() []APIRoute {
	routes := append([]APIRoute{}, PublicAPIRoutes()...)
	return append(routes, PrivateAPIRoutes()...)
}
