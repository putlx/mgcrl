"use strict";

const tasks = new Map();

function lockPage() {
	document.querySelector("body > div").style.visibility = "visible";
	document.querySelector(".main").style.pointerEvents = "none";
	document.querySelector(".main").style.filter = "blur(1px)";
}

function unlockPage() {
	document.querySelector("body > div").style.visibility = "hidden";
	document.querySelector(".main").style.pointerEvents = "auto";
	document.querySelector(".main").style.filter = "";
}

window.addEventListener("load", function () {
	lockPage();

	const socket = new WebSocket(`${window.location.protocol === "https:" ? "wss" : "ws"}://${window.location.host}/downloading`);
	socket.onopen = unlockPage;
	socket.onerror = function (event) {
		toast("连接错误。");
		lockPage();
		console.log(event);
		setTimeout(() => location.reload(), 3000);
	};
	socket.onclose = function (event) {
		toast("连接已断开。");
		lockPage();
		console.log(event);
		setTimeout(() => location.reload(), 3000);
	};
	socket.onmessage = function (event) {
		const message = JSON.parse(event.data);
		const task = tasks.get(message.id);
		if (task) {
			message.element = task.element;
			tasks.set(message.id, message);
			updateTask(message);
		} else {
			appendTask(message);
			updateTask(message);
		}
	};

	const selector = document.querySelector(".control > div.btn-group");
	selector.querySelector("button:nth-child(1)").addEventListener("click", function () {
		document.querySelectorAll("#chapter-list input").forEach(e => e.checked = true);
	});
	selector.querySelector("button:nth-child(2)").addEventListener("click", function () {
		document.querySelectorAll("#chapter-list input").forEach(e => e.checked = false);
	});

	document.querySelector("#url button").addEventListener("click", function () {
		const url = document.querySelector("#url input").value.trim();
		if (!url) return toast("链接不可为空。");

		document.querySelectorAll(".control button").forEach(e => e.disabled = true);
		fetch(location.href, {
			method: "PUT",
			body: JSON.stringify(url),
			headers: new Headers({ "Content-Type": "application/json" })
		})
			.then(response => response.json())
			.then(response => {
				if (typeof response === "string") {
					if (response === "unsupported URL") {
						toast("该链接不受支持。");
					} else {
						toast(`错误：${response}。`);
					}
				} else {
					document.getElementById("chapter-list").innerHTML = response
						.map(chapter => `
							<label class="list-group-item py-1">
								<input class="form-check-input me-1" type="checkbox" value="">
								${chapter.title}
							</label>`)
						.join("");
				}
			})
			.catch(error => {
				toast(`错误：${error.message}。`);
				console.log(error);
			})
			.finally(() => document.querySelectorAll(".control button").forEach(e => e.disabled = false));
	});

	document.querySelector(".control > button").addEventListener("click", function () {
		const indexes = [...document.querySelectorAll("#chapter-list input")]
			.reduce((list, chapter, index) => chapter.checked ? list.push(index) && list : list, []);
		if (indexes.length) {
			document.querySelectorAll(".control button").forEach(e => e.disabled = true);
			fetch(location.href, {
				method: "POST",
				body: JSON.stringify({
					indexes: indexes,
					output: document.getElementById("directory").value.trim()
				}),
				headers: new Headers({ "Content-Type": "application/json" })
			})
				.then(response => response.text())
				.then(response => {
					if (response) {
						toast(`错误：${response}。`);
					}
				})
				.catch(error => {
					toast(`错误：${error.message}。`);
					console.log(error);
				})
				.finally(() => document.querySelectorAll(".control button").forEach(e => e.disabled = false));
		} else {
			toast("没有选中任何章节。");
		}
	});
});

function appendTask(task) {
	const tr = document.createElement("tr");
	tr.innerHTML = `
		<td class="pe-4">${task.manga}</td>
		<td class="pe-4">${task.chapter}</td>
		<td class="pe-4"></td>
		<td class="pe-4"></td>
		<td class="pe-4">
			<div class="progress" style="min-width: 15em;" data-bs-toggle="popover">
				<div role="progressbar"></div>
			</div>
		</td>
		<td class="pe-4">
			<button type="button" class="btn btn-outline-danger btn-sm" disabled>移除</button>
		</td>`;
	task.element = document.querySelector(".tasks > table > tbody").appendChild(tr);
	tr.querySelector("button").onclick = function () {
		fetch(location.href, {
			method: "DELETE",
			body: JSON.stringify(task.id),
			headers: new Headers({ "Content-Type": "application/json" })
		})
			.then(response => response.text())
			.then(response => {
				if (response) {
					toast(`错误：${response}。`);
				} else {
					task.element.remove();
					tasks.delete(task.id);
				}
			})
			.catch(error => {
				toast(`错误：${error.message}。`);
				console.log(error);
			});
	};
	tasks.set(task.id, task);
}

function updateTask(task) {
	task.element.querySelector("td:nth-child(3)").innerHTML = (task.total_size / 1048576).toFixed(2) + " MB";
	task.element.querySelector("td:nth-child(4)").innerHTML = (task.size / 1048576).toFixed(2) + " MB";
	const progressbar = task.element.querySelector("td > div > div");
	if (task.done) {
		progressbar.className = "progress-bar";
		if (task.errors && task.errors.length) {
			progressbar.className += " bg-danger";
			progressbar.setAttribute("data-bs-content", task.errors
				.map(error => (error.filename ? error.filename + ": " : "") + error.error)
				.join("<br>"));
			if (!progressbar.hasAttribute("title")) {
				new bootstrap.Popover(progressbar, { html: true, trigger: "hover", title: "错误日志" });
			}
		} else {
			progressbar.className += " bg-success";
		}
		if (task.total_size !== 0 && task.size <= task.total_size) {
			progressbar.style.width = progressbar.innerHTML =
				(task.size / task.total_size * 100).toFixed(1) + "%";
		} else {
			progressbar.style.width = "75%";
			progressbar.innerHTML = "";
		}
		task.element.querySelector("td > button").disabled = false;
	} else {
		progressbar.className = "progress-bar progress-bar-striped progress-bar-animated bg-info";
		if (task.total_size !== 0 && task.size <= task.total_size) {
			progressbar.style.width = progressbar.innerHTML =
				(task.size / task.total_size * 100).toFixed(1) + "%";
		} else {
			progressbar.className += " indeterminate";
			progressbar.innerHTML = "";
		}
	}
}
