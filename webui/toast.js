window.addEventListener("load", function () {
	const div = this.document.createElement("div");
	div.className = "toast-container";
	div.style.position = "absolute";
	div.style.right = div.style.bottom = "1em";
	div.style.zIndex = "2048";
	document.body.appendChild(div);
});

function toast(message, error = true) {
	let toast = document.createElement("div");
	toast.innerHTML = `
		<div class="toast align-items-center text-white bg-${error ? "danger" : "primary"} border-0" role="alert"
			aria-live="assertive" aria-atomic="true">
			<div class="d-flex">
				<div class="toast-body">
					${message}
				</div>
				<button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"
					aria-label="Close"></button>
			</div>
		</div>`;
	toast = toast.querySelector("div");
	document.querySelector(".toast-container").appendChild(toast);
	new bootstrap.Toast(toast).show();
}
