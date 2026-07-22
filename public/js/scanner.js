function api(method, path, body) {
    return fetch(path, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined
    }).then(function(r) { return r.json(); });
}

window.doScan = function() {
    var s = document.getElementById("scanSubnet").value;
    document.getElementById("scanStatus").innerHTML = '<span class="text-info">Scanning ' + s + '...</span>';
    api("POST", "/api/v1/detect/scan", { subnet: s }).then(function(r) {
        var h = "";
        (r.results || []).forEach(function(d) {
            h += '<div class="col-md-3"><div class="card bg-success bg-opacity-25 border-success"><div class="card-body py-2 text-center"><i class="bi bi-hdd-rack text-success"></i> <strong class="text-light">' + d.ip + '</strong><br><small class="text-success">Port 4370 open</small><br><button class="btn btn-sm btn-outline-light mt-1" onclick="register(\'' + d.ip + '\')">Register</button></div></div></div>';
        });
        document.getElementById("scanResults").innerHTML = h || "<div class=text-secondary>No devices found</div>";
        document.getElementById("scanStatus").innerHTML = '<span class="text-success">Found ' + (r.open || 0) + ' device(s) out of ' + (r.total || 0) + ' scanned</span>';
    }).catch(function() {
        document.getElementById("scanStatus").innerHTML = '<span class="text-danger">Error</span>';
    });
};

window.doDetectOne = function() {
    var ip = document.getElementById("scanOne").value;
    if (!ip) return;
    api("POST", "/api/v1/detect/one", { ip: ip }).then(function(r) {
        alert(r.open ? "OPEN: " + ip + ":4370" : "CLOSED: " + ip + ":4370");
    });
};

window.register = function(ip) {
    var n = prompt("Device name:", ip);
    if (!n) return;
    api("POST", "/api/v1/devices", { name: n, serial_number: "SCAN-" + Date.now(), ip_address: ip }).then(function() {
        alert("Registered: " + n);
    });
};
