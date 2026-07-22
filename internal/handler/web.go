package handler

import (
	"html/template"
	"net/http"
	"path/filepath"
	"time"

	"github.com/jan/goadms/internal/repository"
)

type WebHandler struct {
	deviceRepo     *repository.DeviceRepo
	attendanceRepo *repository.AttendanceRepo
	tmpl           *template.Template
}

func NewWebHandler(deviceRepo *repository.DeviceRepo, attendanceRepo *repository.AttendanceRepo, tmplDir string) *WebHandler {
	layoutPath := filepath.Join(tmplDir, "layout.html")
	tmpl := template.Must(template.New("layout.html").ParseFiles(layoutPath))
	return &WebHandler{deviceRepo: deviceRepo, attendanceRepo: attendanceRepo, tmpl: tmpl}
}

type page struct {
	Title   string
	Content template.HTML
	Scripts template.HTML
}

// --- LoginPage (standalone, no layout) ---
func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html><html lang="en" data-bs-theme="dark"><head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0"><title>ADMS — Login</title><link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css" rel="stylesheet"></head><body class="d-flex align-items-center py-4 bg-dark" style="min-height:100vh"><div class="container" style="max-width:380px"><div class="text-center mb-4"><i class="bi bi-fingerprint text-primary" style="font-size:3rem"></i><h3 class="text-light mt-2">ADMS Login</h3></div><div id="error" class="alert alert-danger d-none"></div><form id="loginForm" onsubmit="doLogin(event)"><div class="mb-3"><label class="form-label text-light">Username</label><input type="text" name="username" class="form-control" required autofocus></div><div class="mb-3"><label class="form-label text-light">Password</label><input type="password" name="password" class="form-control" required></div><button type="submit" class="btn btn-primary w-100">Login</button></form></div><script>async function doLogin(e){e.preventDefault();const f=new FormData(document.getElementById("loginForm"));try{const r=await fetch("/api/v1/auth/login",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({username:f.get("username"),password:f.get("password")})});const j=await r.json();if(!r.ok)throw new Error(j.error||"Login failed");localStorage.setItem("access_token",j.access_token);localStorage.setItem("refresh_token",j.refresh_token);window.location.href="/web/dashboard"}catch(e){document.getElementById("error").textContent=e.message;document.getElementById("error").classList.remove("d-none")}}</script></body></html>`))
}

// --- Dashboard ---
func (h *WebHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<h2 class="text-light mb-4">Dashboard</h2><div class="row g-3 mb-4" id="statCards"><div class="col-md-3"><div class="card bg-primary bg-opacity-25 border-primary"><div class="card-body text-center"><h1 id="statTotal" class="text-primary">—</h1><small class="text-light">Total Devices</small></div></div></div><div class="col-md-3"><div class="card bg-success bg-opacity-25 border-success"><div class="card-body text-center"><h1 id="statOnline" class="text-success">—</h1><small class="text-light">Online</small></div></div></div><div class="col-md-3"><div class="card bg-danger bg-opacity-25 border-danger"><div class="card-body text-center"><h1 id="statOffline" class="text-danger">—</h1><small class="text-light">Offline</small></div></div></div><div class="col-md-3"><div class="card bg-info bg-opacity-25 border-info"><div class="card-body text-center"><h1 id="statToday" class="text-info">—</h1><small class="text-light">Today Records</small></div></div></div></div><h4 class="text-light mt-4">Device Status</h4><div class="row g-3" id="deviceCards"><div class="text-secondary">Loading...</div></div><h4 class="text-light mt-4">Live Events</h4><div id="liveEvents" class="border border-secondary rounded p-2" style="max-height:200px;overflow-y:auto;font-size:.8rem"><span class="text-secondary">Connecting...</span></div>`)
	scripts := template.HTML(`<script src="/public/js/dashboard.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Dashboard", Content: content, Scripts: scripts})
}

// --- Devices ---
func (h *WebHandler) DevicesPage(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<div class="d-flex justify-content-between align-items-center mb-3"><h2 class="text-light">Devices</h2><button class="btn btn-primary btn-sm" data-bs-toggle="modal" data-bs-target="#addDeviceModal" onclick="document.getElementById('modalTitle').textContent='Add Device';document.getElementById('saveBtn').textContent='Save';document.getElementById('addDeviceForm').reset()"><i class="bi bi-plus-lg"></i> Add Device</button></div><div class="table-responsive"><table class="table table-dark table-hover"><thead><tr><th>Name</th><th>Serial Number</th><th>IP</th><th>Location</th><th>Status</th><th>Last Seen</th><th>Actions</th></tr></thead><tbody id="devTable"><tr><td colspan="7" class="text-center text-secondary">Loading...</td></tr></tbody></table></div><div class="modal fade" id="addDeviceModal" tabindex="-1"><div class="modal-dialog"><div class="modal-content bg-dark-subtle"><div class="modal-header"><h5 class="modal-title text-light" id="modalTitle">Add Device</h5><button type="button" class="btn-close" data-bs-dismiss="modal"></button></div><div class="modal-body"><form id="addDeviceForm" onsubmit="Devices.add(event)"><div class="mb-2"><label class="form-label text-light">Name *</label><input type="text" name="name" id="devName" class="form-control" required></div><div class="mb-2"><label class="form-label text-light">Serial Number *</label><input type="text" name="serial_number" id="devSN" class="form-control" required></div><div class="mb-2"><label class="form-label text-light">IP Address</label><input type="text" name="ip_address" id="devIP" class="form-control"></div><div class="mb-2"><label class="form-label text-light">Port</label><input type="number" name="port" id="devPort" class="form-control" value="4370"></div><div class="mb-2"><label class="form-label text-light">Location</label><input type="text" name="location" id="devLoc" class="form-control"></div><button type="submit" class="btn btn-primary w-100" id="saveBtn">Save</button></form></div></div></div></div>`)
	scripts := template.HTML(`<script src="/public/js/devices.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Devices", Content: content, Scripts: scripts})
}

// --- Attendance ---
func (h *WebHandler) AttendancePage(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")
	content := template.HTML(`<h2 class="text-light mb-3">Attendance Log</h2><div class="row g-2 mb-3"><div class="col-auto"><select id="filterDevice" class="form-select form-select-sm" onchange="loadAttendance()"><option value="">All Devices</option></select></div><div class="col-auto"><input type="date" id="filterDate" class="form-control form-control-sm" onchange="loadAttendance()" value="` + template.HTML(today) + `"></div><div class="col-auto"><input type="text" id="filterEmp" class="form-control form-control-sm" placeholder="Employee ID" onchange="loadAttendance()"></div><div class="col-auto"><button class="btn btn-sm btn-outline-secondary" onclick="loadAttendance()"><i class="bi bi-arrow-clockwise"></i> Refresh</button></div></div><div class="table-responsive"><table class="table table-dark table-hover table-sm"><thead><tr><th>Device</th><th>Employee ID</th><th>Timestamp</th><th>Status</th><th>Source</th></tr></thead><tbody id="attTable"><tr><td colspan="5" class="text-center text-secondary">Loading...</td></tr></tbody></table></div><div id="attInfo" class="text-secondary small"></div>`)
	scripts := template.HTML(`<script src="/public/js/attendance.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Attendance", Content: content, Scripts: scripts})
}

// --- Shifts ---
func (h *WebHandler) ShiftsPage(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<h2 class="text-light mb-3">Shifts</h2><div class="row g-3"><div class="col-md-5"><div class="card bg-dark-subtle"><div class="card-body"><h5 class="text-light">Add Shift</h5><form id="shiftForm" onsubmit="Shifts.add(event)"><div class="mb-2"><input name="name" class="form-control form-control-sm" placeholder="Shift name" required></div><div class="row g-2"><div class="col"><input name="start_time" class="form-control form-control-sm" placeholder="Start (08:00)" value="08:00"></div><div class="col"><input name="end_time" class="form-control form-control-sm" placeholder="End (17:00)" value="17:00"></div></div><div class="row g-2 mt-1"><div class="col"><input name="late_tolerance_minutes" type="number" class="form-control form-control-sm" placeholder="Late tol (min)" value="5"></div><div class="col"><input name="overtime_after_minutes" type="number" class="form-control form-control-sm" placeholder="OT after (min)" value="0"></div></div><button class="btn btn-primary btn-sm w-100 mt-2">Save</button></form></div></div></div><div class="col-md-7"><div id="shiftList"></div></div></div>`)
	scripts := template.HTML(`<script src="/public/js/shifts.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Shifts", Content: content, Scripts: scripts})
}

// --- Reports ---
func (h *WebHandler) ReportsPage(w http.ResponseWriter, r *http.Request) {
	today := time.Now().Format("2006-01-02")
	content := template.HTML(`<h2 class="text-light mb-3">Attendance Report</h2><div class="row g-2 mb-3"><div class="col-auto"><input type="date" id="rDateFrom" class="form-control form-control-sm" value="` + template.HTML(today) + `"></div><div class="col-auto"><input type="date" id="rDateTo" class="form-control form-control-sm" value="` + template.HTML(today) + `"></div><div class="col-auto"><select id="rDevice" class="form-select form-select-sm"><option value="">All Devices</option></select></div><div class="col-auto"><button class="btn btn-sm btn-primary" onclick="loadReport()"><i class="bi bi-search"></i> Generate</button></div><div class="col-auto"><button class="btn btn-sm btn-success" onclick="downloadCSV()"><i class="bi bi-download"></i> CSV</button></div></div><div class="row g-3 mb-3" id="reportSummary"></div><div class="table-responsive" style="max-height:60vh;overflow-y:auto"><table class="table table-dark table-sm"><thead><tr><th>Date</th><th>Employee</th><th>Device</th><th>Check In</th><th>Check Out</th><th>Late</th><th>Early</th><th>OT</th><th>Status</th></tr></thead><tbody id="reportBody"><tr><td colspan="9" class="text-center text-secondary">Click Generate</td></tr></tbody></table></div>`)
	scripts := template.HTML(`<script src="/public/js/reports.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Reports", Content: content, Scripts: scripts})
}

// --- Fingerprint Users ---
func (h *WebHandler) FingerUsersPage(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<h2 class="text-light mb-3">Fingerprint Users</h2><div class="row g-2 mb-3"><div class="col-auto"><select id="fuDevice" class="form-select form-select-sm"><option value="">Select target device</option></select></div><div class="col-auto"><button class="btn btn-sm btn-success" onclick="Users.syncAll()"><i class="bi bi-upload"></i> Sync All Users</button></div></div><div class="row g-3"><div class="col-md-4"><div class="card bg-dark-subtle"><div class="card-body"><h5 class="text-light">Add User</h5><form id="userForm" onsubmit="Users.add(event)"><div class="mb-2"><input name="employee_code" class="form-control form-control-sm" placeholder="Employee Code" required></div><div class="mb-2"><input name="full_name" class="form-control form-control-sm" placeholder="Full Name" required></div><div class="mb-2"><input name="department" class="form-control form-control-sm" placeholder="Department"></div><button class="btn btn-primary btn-sm w-100">Save</button></form></div></div></div><div class="col-md-8"><div id="userList"></div></div></div>`)
	scripts := template.HTML(`<script src="/public/js/users.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Users", Content: content, Scripts: scripts})
}

// --- Webhooks ---
func (h *WebHandler) WebhooksPage(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<h2 class="text-light mb-3">Webhooks</h2><div class="row g-3"><div class="col-md-3"><div class="card bg-dark-subtle"><div class="card-body"><h5 class="text-light">Select Device</h5><select id="whDevice" class="form-select form-select-sm" onchange="Webhooks.load()"><option value="">— Select Device —</option></select></div></div><div class="card bg-dark-subtle mt-2"><div class="card-body"><h5 class="text-light">Add Webhook</h5><form id="whForm" onsubmit="Webhooks.add(event)"><div class="mb-2"><input name="name" class="form-control form-control-sm" placeholder="Name" required></div><div class="mb-2"><input name="url" class="form-control form-control-sm" placeholder="URL (https://...)" required></div><div class="mb-2"><input name="secret" class="form-control form-control-sm" placeholder="Secret (optional)"></div><div class="mb-2"><label class="form-label text-light small">Events:</label><select name="events" class="form-select form-select-sm" multiple size="4"><option value="attendance.created">attendance.created</option><option value="device.online">device.online</option><option value="device.offline">device.offline</option><option value="handshake.received">handshake.received</option><option value="pull.completed">pull.completed</option><option value="pull.failed">pull.failed</option></select></div><button class="btn btn-primary btn-sm w-100">Save</button></form></div></div></div><div class="col-md-9"><div id="whList"><div class="text-secondary">Select a device to see webhooks</div></div></div></div>`)
	scripts := template.HTML(`<script src="/public/js/webhooks.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Webhooks", Content: content, Scripts: scripts})
}

// --- Scanner ---
func (h *WebHandler) ScannerPage(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<h2 class="text-light mb-3">Device Scanner</h2><div class="row g-2 mb-3"><div class="col-auto"><input id="scanSubnet" class="form-control form-control-sm" placeholder="Subnet" value="192.168.1.0/24" style="width:220px"></div><div class="col-auto"><button class="btn btn-sm btn-primary" onclick="doScan()"><i class="bi bi-search"></i> Scan</button></div><div class="col-auto"><input id="scanOne" class="form-control form-control-sm" placeholder="Single IP" style="width:160px"></div><div class="col-auto"><button class="btn btn-sm btn-outline-info" onclick="doDetectOne()"><i class="bi bi-pin"></i> Detect</button></div></div><div id="scanStatus" class="mb-2"></div><div id="scanResults" class="row g-2"></div>`)
	scripts := template.HTML(`<script src="/public/js/scanner.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Scanner", Content: content, Scripts: scripts})
}

// --- Device Users ---
func (h *WebHandler) DeviceUsersPage(w http.ResponseWriter, r *http.Request) {
	content := template.HTML(`<h2 class="text-light mb-3">Device Users</h2><div class="row g-2 mb-3"><div class="col-auto"><select id="duDevice" class="form-select form-select-sm" onchange="loadDeviceUsers()"><option value="">— Select Device —</option></select></div><div class="col-auto"><button class="btn btn-sm btn-primary" onclick="loadDeviceUsers()"><i class="bi bi-arrow-clockwise"></i> Refresh</button></div></div><div id="devUserList"><div class="text-secondary">Select a device to view users</div></div>`)
	scripts := template.HTML(`<script src="/public/js/device-users.js"></script>`)
	h.tmpl.Execute(w, &page{Title: "Device Users", Content: content, Scripts: scripts})
}
