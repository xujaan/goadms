(function() {
    function api(method, path, body) {
        return fetch(path, {
            method: method,
            headers: { 'Content-Type': 'application/json' },
            body: body ? JSON.stringify(body) : undefined
        }).then(function(r) { return r.json(); });
    }

    async function load() {
        try {
            var d = await api("GET", "/api/v1/shifts");
            var shifts = d.data || [];
            var h = "";
            shifts.forEach(function(s) {
                var nameEsc = s.name.replace(/'/g, "\\'");
                h += '<div class="card bg-dark-subtle mb-2"><div class="card-body py-2"><div class="d-flex justify-content-between">' +
                    '<div><strong class="text-light">' + s.name + '</strong> ' +
                    '<small class="text-secondary">' + s.start_time + "-" + s.end_time +
                    ' | late:' + s.late_tolerance_minutes + 'm | OT:' + s.overtime_after_minutes + 'm</small><br>' +
                    '<small>' + (s.is_active ? '<span class="text-success">Active</span>' : '<span class="text-danger">Inactive</span>') + '</small></div>' +
                    '<div>' +
                    '<button class="btn btn-sm btn-outline-secondary me-1" onclick="Shifts.edit(\'' + s.id + '\',\'' + nameEsc + '\',\'' + s.start_time + '\',\'' + s.end_time + '\')"><i class="bi bi-pencil"></i></button>' +
                    '<button class="btn btn-sm btn-outline-danger" onclick="Shifts.del(\'' + s.id + '\')"><i class="bi bi-trash"></i></button>' +
                    '</div></div></div></div>';
            });
            document.getElementById("shiftList").innerHTML = h || '<div class="text-secondary">No shifts</div>';
        } catch (e) { console.error(e); }
    }

    window.Shifts = {
        load: load,
        add: async function(e) {
            e.preventDefault();
            var f = document.getElementById("shiftForm");
            var d = {};
            new FormData(f).forEach(function(v, k) { d[k] = v; });
            d.late_tolerance_minutes = parseInt(d.late_tolerance_minutes) || 5;
            d.overtime_after_minutes = parseInt(d.overtime_after_minutes) || 0;
            await api("POST", "/api/v1/shifts", d);
            f.reset();
            load();
        },
        edit: async function(id, name, st, et) {
            var nn = prompt("Name", name);
            if (!nn) return;
            var sst = prompt("Start time (08:00)", st);
            var eet = prompt("End time (17:00)", et);
            await api("PUT", "/api/v1/shifts/" + id, { name: nn, start_time: sst, end_time: eet });
            load();
        },
        del: async function(id) {
            if (!confirm("Delete shift?")) return;
            await api("DELETE", "/api/v1/shifts/" + id);
            load();
        }
    };

    load();
})();
