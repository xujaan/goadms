function api(method, path, body) {
    return fetch(path, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: body ? JSON.stringify(body) : undefined
    }).then(function(r) { return r.json(); });
}

window.loadDeviceUsers = function() {
    var id = document.getElementById("duDevice").value;
    if (!id) {
        document.getElementById("devUserList").innerHTML = '<div class="text-secondary">Select a device</div>';
        return;
    }
    document.getElementById("devUserList").innerHTML = '<div class="text-info">Connecting via TCP...</div>';
    api("GET", "/api/v1/devices/" + id + "/users").then(function(d) {
        var u = d.data || [];
        var h = '<div class="mb-2"><strong class="text-light">' + u.length + ' users on device</strong></div>';
        u.forEach(function(u) {
            h += '<div class="card bg-dark-subtle mb-1"><div class="card-body py-2"><div class="d-flex justify-content-between">' +
                '<div><strong class="text-light">' + (u.name || "User " + u.user_id) + '</strong> <code class="text-secondary">UID:' + u.user_id + '</code><br>' +
                '<small class="text-secondary">Card: ' + (u.card_no || "—") + ' | Priv:' + u.privilege + '</small></div>' +
                '<div><button class="btn btn-sm btn-outline-danger" onclick="delDeviceUser(\'' + u.user_id + '\')">Del</button></div>' +
                '</div></div></div>';
        });
        document.getElementById("devUserList").innerHTML = h || "<div class=text-secondary>No users</div>";
    }).catch(function(e) {
        document.getElementById("devUserList").innerHTML = '<div class="text-danger">Failed: ' + e.message + '</div>';
    });
};

window.delDeviceUser = function(uid) {
    if (!confirm("Delete user " + uid + " from device?")) return;
    var did = document.getElementById("duDevice").value;
    api("POST", "/api/v1/devices/" + did + "/users/" + uid + "/delete", {}).then(loadDeviceUsers);
};

// Load device dropdown.
fetch("/api/v1/devices").then(function(r) { return r.json(); }).then(function(d) {
    (d.data || []).forEach(function(d) {
        document.getElementById("duDevice").innerHTML += '<option value="' + d.id + '">' + d.name + ' (' + d.ip_address + ':' + d.port + ')</option>';
    });
});
