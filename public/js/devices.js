(function() {
    var loadTimer, editingId = null;

    function api(method, path, body) {
        return fetch(path, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: body ? JSON.stringify(body) : undefined
        }).then(function(r) { if (!r.ok) throw new Error(r.status); return r.json(); });
    }

    function row(d) {
        var sc = d.is_online ? "success" : "danger";
        var st = d.is_online ? "Online" : "Offline";
        var ls = d.last_handshake_at ? new Date(d.last_handshake_at).toLocaleString() : "Never";
        var ip = (d.ip_address || "—") + ":" + d.port;
        var loc = d.location || "—";
        var nm = esc(d.name), ipE = esc(d.ip_address||""), locE = esc(d.location||"");

        return '<tr>' +
            '<td>' + d.name + '</td><td><code>' + d.serial_number + '</code></td><td>' + ip + '</td><td>' + loc + '</td>' +
            '<td><span class="badge bg-' + sc + '">' + st + '</span></td><td>' + ls + '</td>' +
            '<td>' +
            '<button class="btn btn-sm btn-outline-secondary me-1" title="Edit" onclick="Devices.edit(\'' + d.id + '\',\'' + nm + '\',\'' + ipE + '\',\'' + locE + '\',\'' + (d.serial_number||"") + '\')"><i class="bi bi-pencil"></i></button>' +
            '<button class="btn btn-sm btn-outline-info me-1" title="Pull" onclick="Devices.pull(\'' + d.id + '\')"><i class="bi bi-download"></i></button>' +
            '<button class="btn btn-sm btn-outline-warning" title="Delete" onclick="Devices.del(\'' + d.id + '\',\'' + nm + '\')"><i class="bi bi-trash"></i></button>' +
            '</td></tr>';
    }

    async function load() {
        try {
            var d = await api("GET", "/api/v1/devices");
            var devs = d.data || [];
            document.getElementById("devTable").innerHTML = devs.map(row).join("") || '<tr><td colspan="7" class="text-center text-secondary">No devices</td></tr>';
        } catch (e) {
            document.getElementById("devTable").innerHTML = '<tr><td colspan="7" class="text-danger">Failed</td></tr>';
        }
    }

    function esc(s) { return (s||"").replace(/'/g,"\\'").replace(/"/g,"&quot;"); }

    window.Devices = {
        load: load,
        pull: async function(id) {
            try { var r = await api("POST", "/api/v1/devices/" + id + "/pull-attendance"); alert("Records: " + (r.records_pulled || 0)); } catch(e) { alert("Failed: " + e.message); }
            load();
        },
        del: async function(id, name) {
            if (!confirm("Delete " + name + "?")) return;
            await api("DELETE", "/api/v1/devices/" + id);
            load();
        },
        edit: function(id, name, ip, loc, sn) {
            editingId = id;
            document.getElementById("modalTitle").textContent = "Edit Device";
            document.getElementById("devName").value = name;
            document.getElementById("devSN").value = sn;
            document.getElementById("devIP").value = ip;
            document.getElementById("devLoc").value = loc;
            document.getElementById("devPort").value = 4370;
            document.getElementById("saveBtn").textContent = "Update";
            new bootstrap.Modal(document.getElementById("addDeviceModal")).show();
        },
        add: async function(e) {
            e.preventDefault();
            var f = document.getElementById("addDeviceForm");
            var d = {}; new FormData(f).forEach(function(v, k) { d[k] = v; });
            d.port = parseInt(d.port) || 4370;
            if (editingId) {
                await api("PUT", "/api/v1/devices/" + editingId, d);
                editingId = null;
            } else {
                await api("POST", "/api/v1/devices", d);
            }
            document.getElementById("saveBtn").textContent = "Save";
            document.getElementById("modalTitle").textContent = "Add Device";
            f.reset();
            var mc = document.querySelector("#addDeviceModal .btn-close");
            if (mc) mc.click();
            load();
        }
    };

    // Reset modal on close.
    document.getElementById("addDeviceModal").addEventListener("hidden.bs.modal", function() {
        editingId = null;
        document.getElementById("saveBtn").textContent = "Save";
        document.getElementById("modalTitle").textContent = "Add Device";
        document.getElementById("addDeviceForm").reset();
    });

    load();
    loadTimer = setInterval(load, 60000);
})();
