import json
import os
import subprocess
import sys
import tkinter as tk
import tkinter.filedialog
import tkinter.messagebox as messagebox
import tkinter.scrolledtext
from threading import Lock, Thread, Event
from tkinter import ttk

app = tk.Tk()
app.title("mgcrl")
app.geometry("1000x500")
res_path = getattr(sys, "_MEIPASS", os.path.dirname(os.path.abspath(__file__)))
icon = os.path.join(res_path, "icon.ico")
app.iconbitmap(icon, icon)
app.grid_columnconfigure(0, weight=3, uniform="k")
app.grid_columnconfigure(1, weight=5, uniform="k")
app.grid_rowconfigure(0, weight=1)

frame = ttk.Frame(app)
frame.grid(row=0, column=0, sticky=tk.W + tk.E + tk.N + tk.S)
frame.grid_columnconfigure(1, weight=1)
frame.grid_rowconfigure(1, weight=1)

ttk.Label(frame, text="链接：").grid(row=0, column=0)
url = tk.StringVar(frame)
link = ttk.Entry(frame, textvariable=url)
link.grid(row=0, column=1, sticky=tk.W + tk.E)
fetch = ttk.Button(frame, text="获取", command=lambda: Thread(target=get_catalog).start())
fetch.grid(row=0, column=2)
link.bind("<Key-Return>", lambda _: fetch.invoke())

box = ttk.Frame(frame)
box.grid(row=1, column=0, columnspan=3, sticky=tk.W + tk.E + tk.N + tk.S)
itemlist = ttk.Treeview(box, columns=("t",), show="headings", selectmode=tk.EXTENDED)
scrollbar = ttk.Scrollbar(box, command=itemlist.yview)
scrollbar.pack(side=tk.RIGHT, fill=tk.BOTH)
itemlist.heading("t", text="标题")
itemlist.pack(fill=tk.BOTH, expand=True)
itemlist.config(yscrollcommand=scrollbar.set)

ttk.Label(frame, text="目录：").grid(row=2, column=0)
output = tk.StringVar(frame, value=os.getcwd())
ttk.Entry(frame, textvariable=output).grid(row=2, column=1, sticky=tk.W + tk.E)
browse = ttk.Button(frame, text="浏览")
browse.grid(row=2, column=2)
browse.config(command=lambda: output.set(tk.filedialog.askdirectory() or output.get()))
dl = ttk.Button(frame, text="下载选中的章节", command=lambda: Thread(target=download).start())
dl.grid(row=3, column=0, columnspan=3, sticky=tk.W + tk.E)

frame = ttk.Frame(app)
frame.grid(row=0, column=1, sticky=tk.W + tk.E + tk.N + tk.S)
frame.grid_columnconfigure(0, weight=1)
frame.grid_rowconfigure(0, weight=7, uniform="p")
frame.grid_rowconfigure(1, weight=2, uniform="p")

box = ttk.Frame(frame)
box.grid(row=0, column=0, sticky=tk.W + tk.E + tk.N + tk.S)
tasklist = ttk.Treeview(box, columns=("m", "c", "t", "p", "s"), show="headings", selectmode=tk.BROWSE)
scrollbar = ttk.Scrollbar(box, command=tasklist.yview, orient=tk.VERTICAL)
scrollbar.pack(side=tk.RIGHT, fill=tk.BOTH)
tasklist.config(yscrollcommand=scrollbar.set)
tasklist.column("m", width=12)
tasklist.column("c", width=12)
tasklist.column("t", width=12)
tasklist.column("p", width=12)
tasklist.column("s", width=12)
tasklist.heading("m", text="漫画标题")
tasklist.heading("c", text="章节标题")
tasklist.heading("t", text="状态")
tasklist.heading("p", text="进度")
tasklist.heading("s", text="大小")
tasklist.pack(fill=tk.BOTH, expand=True)
tasklist.bind("<Button-1>", lambda _: evt.set())

log = tk.scrolledtext.ScrolledText(frame, wrap=tk.WORD, state=tk.DISABLED)
log.grid(row=1, column=0, sticky=tk.W + tk.E + tk.N + tk.S)

for frame in app.winfo_children():
    frame.grid_configure(padx=4, pady=4)
    for child in frame.winfo_children():
        child.grid_configure(padx=3, pady=3)
link.focus()


class control_panel:
    def __enter__(self):
        fetch.config(state=tk.DISABLED)
        dl.config(state=tk.DISABLED)
        browse.config(state=tk.DISABLED)

    def __exit__(self, exc_type, exc_value, traceback):
        fetch.config(state=tk.NORMAL)
        dl.config(state=tk.NORMAL)
        browse.config(state=tk.NORMAL)


control_panel = control_panel()

if hasattr(subprocess, "STARTUPINFO"):
    si = subprocess.STARTUPINFO()
    si.dwFlags |= subprocess.STARTF_USESHOWWINDOW
else:
    si = None
proc = subprocess.Popen(
    [os.path.join(res_path, "repl")],
    stdin=subprocess.PIPE,
    stdout=subprocess.PIPE,
    stderr=subprocess.PIPE,
    startupinfo=si,
)
lock = Lock()
evt = Event()
tasks = []
dump = None


def get_catalog():
    with control_panel:
        URL = url.get().strip()
        if len(URL) == 0:
            messagebox.showwarning("警告", "请输入链接。")
            return

        proc.stdin.write(json.dumps({"url": URL}).encode())
        proc.stdin.write(b"\n")
        proc.stdin.flush()
        data = json.loads(proc.stdout.readline())
        if isinstance(data, str):
            if data == "unsupported URL":
                messagebox.showerror("错误", "该链接不受支持。")
            else:
                messagebox.showerror("错误", f"错误：{data}。")
            return
        itemlist.delete(*itemlist.get_children())
        itemlist.heading("t", text=data["title"])
        for c in data["chapters"]:
            itemlist.insert("", tk.END, values=(c["title"],))
        global dump
        dump = data
        dump["url"] = URL


def download():
    with control_panel:
        if len(itemlist.selection()) == 0:
            messagebox.showwarning("警告", "没有选中任何章节。")
            return

        idxes = [itemlist.index(n) for n in itemlist.selection()]
        data = {
            "url": dump["url"],
            "output": output.get().strip(),
            "indexes": idxes,
        }
        proc.stdin.write(json.dumps(data).encode())
        proc.stdin.write(b"\n")
        proc.stdin.flush()

        with lock:
            for idx in idxes:
                data = [dump["title"], dump["chapters"][idx]["title"], "下载中", "?", "?"]
                tasklist.insert("", tk.END, values=data)
                tasks.append(data)

            proc.stdout.readline()


def update():
    while proc.poll() is None:
        s = proc.stderr.readline().rstrip()
        if len(s) == 0:
            break
        s = json.loads(s)
        with lock:
            data = tasks[s["index"]]
            data[4] = f"{s['total_size'] / 1048576:.1f} MB"
            if s["total_size"] == 0 or s["size"] > s["total_size"]:
                data[3] = "?"
            else:
                data[3] = f"{100 * s['size'] // s['total_size']} %"
            if s["errors"] is not None:
                errmsg = ""
                data[2] = "已完成" if len(s["errors"]) == 0 else "出错"
                for e in s["errors"]:
                    if len(e["filename"]) != 0:
                        errmsg += e["filename"] + ": "
                    errmsg += e["error"] + "\n"
            tasklist.item(tasklist.get_children()[s["index"]], values=data)
            if s["errors"] is not None:
                tasks[s["index"]] = errmsg


def show_log():
    while evt.wait() and proc.poll() is None:
        evt.clear()
        items = tasklist.selection()
        if len(items) != 0:
            log.config(state=tk.NORMAL)
            log.delete(1.0, tk.END)
            with lock:
                data = tasks[tasklist.index(items[0])]
            if isinstance(data, str):
                log.insert(tk.END, data)
            log.config(state=tk.DISABLED)


Thread(target=update).start()
Thread(target=show_log).start()
app.mainloop()
proc.kill()
proc.wait()
evt.set()
