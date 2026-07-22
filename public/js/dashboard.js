(function() {
    function api(path) {
        return fetch(path).then(function(r) { return r.json(); });
    }

    async function load() {
        try {
            var d = await api("/api/v1/devices");
            var devs = d.data || [];
            var online = 0, offline = 0;
            devs.forEach(function(d) { d.is_online ? online++ : offline++; });
            document.getElementById("statTotal").textContent = devs.length;
            document.getElementById("statOnline").textContent = online;
            document.getElementById("statOffline").textContent = offline;

            var h = "";
            devs.forEach(function(d) {
                var icon = d.is_online ? '<i class="bi bi-circle-fill text-success"></i>' : '<i class="bi bi-circle-fill text-danger"></i>';
                h += '<div class="col-md-4"><div class="card device-card"><div class="card-body">' +
                    '<h5>' + icon + ' ' + d.name + '</h5>' +
                    '<small class="text-secondary">SN: ' + d.serial_number + '<br>IP: ' + (d.ip_address || "—") + ":" + d.port + '</small><br>' +
                    '<small class="' + (d.is_online ? 'text-success' : 'text-danger') + '">' + (d.is_online ? 'Online' : 'Offline') + '</small>' +
                    '</div></div></div>';
            });
            document.getElementById("deviceCards").innerHTML = h || '<div class="text-secondary">No devices</div>';

            var today = new Date().toISOString().slice(0, 10);
            var a = await api("/api/v1/attendances?date_from=" + today + "&limit=1");
            document.getElementById("statToday").textContent = a.total || 0;
        } catch (e) {
            document.getElementById("deviceCards").innerHTML = '<div class="text-danger">Failed to load</div>';
        }
    }

    // SSE live events
    var es = new EventSource("/api/v1/events");
    var el = document.getElementById("liveEvents");
    es.addEventListener("connected", function(e) {
        el.innerHTML = '<span class="text-success">● Live</span><br>';
    });
    es.addEventListener("attendance", function(e) {
        var d = JSON.parse(e.data);
        el.innerHTML += '<div><span class="text-info">' + new Date(d.timestamp).toLocaleTimeString() +
            '</span> — ' + d.employee_id + ' @ ' + d.device_sn + '</div>';
        el.scrollTop = el.scrollHeight;
    });

    load();
    setInterval(load, 30000);
})();
