(function() {
    function loadDevices() {
        fetch("/api/v1/devices").then(function(r) { return r.json(); }).then(function(d) {
            var sel = document.getElementById("filterDevice");
            (d.data || []).forEach(function(d) {
                sel.innerHTML += '<option value="' + d.serial_number + '">' + d.name + '</option>';
            });
        });
    }

    window.loadAttendance = function() {
        var sn = document.getElementById("filterDevice").value;
        var date = document.getElementById("filterDate").value;
        var emp = document.getElementById("filterEmp").value;
        var url = "/api/v1/attendances?limit=100";
        if (sn) url += "&device_sn=" + encodeURIComponent(sn);
        if (date) url += "&date_from=" + date;
        if (emp) url += "&employee_id=" + emp;

        fetch(url).then(function(r) { return r.json(); }).then(function(data) {
            var records = data.data || [];
            var h = "";
            records.forEach(function(a) {
                h += '<tr><td><code>' + a.device_sn + '</code></td>' +
                    '<td>' + a.employee_id + '</td>' +
                    '<td>' + new Date(a.timestamp).toLocaleString() + '</td>' +
                    '<td>' + a.status1 + '|' + a.status2 + '|' + a.status3 + '</td>' +
                    '<td><span class="badge bg-secondary">' + a.source + '</span></td></tr>';
            });
            document.getElementById("attTable").innerHTML = h || '<tr><td colspan="5" class="text-center text-secondary">No records</td></tr>';
            document.getElementById("attInfo").textContent = 'Showing ' + records.length + ' of ' + (data.total || 0) + ' records';
        }).catch(function() {
            document.getElementById("attTable").innerHTML = '<tr><td colspan="5" class="text-danger">Failed to load</td></tr>';
        });
    };

    document.getElementById("filterDate").value = new Date().toISOString().slice(0, 10);
    loadDevices();
    loadAttendance();
})();
