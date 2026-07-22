var lastReport = null;

function api(p) { return fetch(p).then(function(r) { return r.json(); }); }

window.loadReport = function() {
    var f = document.getElementById("rDateFrom").value;
    var d = document.getElementById("rDateTo").value;
    var s = document.getElementById("rDevice").value;
    var u = "/api/v1/reports/attendance?date_from=" + f + "&date_to=" + d;
    if (s) u += "&device_sn=" + s;
    api(u).then(function(r) {
        lastReport = r;
        var h = "";
        (r.reports || []).forEach(function(a) {
            h += "<tr><td>" + a.date + "</td><td>" + a.employee_id + "</td><td>" + a.device_sn + "</td><td>" + a.check_in + "</td><td>" + a.check_out + "</td><td>" + a.late_minutes + "</td><td>" + a.early_minutes + "</td><td>" + a.overtime_minutes + "</td><td><span class=\"badge bg-" + statusColor(a.status) + "\">" + a.status + "</span></td></tr>";
        });
        document.getElementById("reportBody").innerHTML = h || "<tr><td colspan=9 class=text-center>No data</td></tr>";
        var s = r.summary;
        document.getElementById("reportSummary").innerHTML =
            '<div class=col-md-3><div class="card bg-dark-subtle"><div class="card-body text-center"><h4>' + s.total_records + '</h4><small>Total</small></div></div></div>' +
            '<div class=col-md-3><div class="card bg-success bg-opacity-25"><div class="card-body text-center"><h4>' + s.hadir + '</h4><small>Hadir</small></div></div></div>' +
            '<div class=col-md-3><div class="card bg-warning bg-opacity-25"><div class="card-body text-center"><h4>' + s.terlambat + '</h4><small>Terlambat</small></div></div></div>';
    }).catch(function(e) { alert("Failed: " + e.message); });
};

window.downloadCSV = function() {
    var f = document.getElementById("rDateFrom").value;
    var d = document.getElementById("rDateTo").value;
    var s = document.getElementById("rDevice").value;
    var u = "/api/v1/reports/attendance.csv?date_from=" + f + "&date_to=" + d;
    if (s) u += "&device_sn=" + s;
    window.open(u, "_blank");
};

function statusColor(s) {
    if (s === "Hadir") return "success";
    if (s === "Terlambat") return "warning";
    if (s === "Pulang Dulu") return "info";
    if (s === "Lembur") return "primary";
    return "secondary";
}

// Load device filter dropdown.
fetch("/api/v1/devices").then(function(r) { return r.json(); }).then(function(d) {
    (d.data || []).forEach(function(d) {
        document.getElementById("rDevice").innerHTML += '<option value="' + d.serial_number + '">' + d.name + '</option>';
    });
});
