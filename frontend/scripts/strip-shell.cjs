const fs = require("fs");
const path = require("path");

const root = path.join(__dirname, "..", "src", "app", "(app)");
const targets = [];
(function walk(dir) {
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) walk(p);
    else if (e.name === "page.tsx") targets.push(p);
  }
})(root);

const stripShell = (src) => {
  let out = src;
  const before = out;
  // 1. Remove <SidebarProvider ...> <AppSidebar /> <SidebarInset> opening block
  out = out.replace(/<SidebarProvider\b[\s\S]*?>\s*\n\s*<AppSidebar \/>\s*\n\s*<SidebarInset>\s*\n/, "");
  if (out === before) return null;

  // 2. Replace header opening + everything up to ml-auto div with content wrapper + toolbar
  out = out.replace(
    /<header className="sticky top-0 flex h-14 shrink-0 items-center gap-2 border-b bg-background px-4">[\s\S]*?<div className="flex items-center gap-2 ml-auto">/,
    '<div className="flex flex-1 flex-col gap-6 p-6">\n        <div className="flex flex-wrap items-center gap-3">'
  );

  // 3. Remove NotificationsPopover from the page (now lives in (app) layout header)
  out = out.replace(/\s*<NotificationsPopover \/>\s*\n/, "\n");

  // 4. Remove closing </header> + the now-duplicated original content wrapper opening
  out = out.replace(/\n\s*<\/header>\n\s*<div className="flex flex-1 flex-col gap-6 p-6">\n/, "\n");

  // 5. Remove </SidebarInset> + </SidebarProvider> closings
  out = out.replace(/\s*<\/SidebarInset>\s*\n\s*<\/SidebarProvider>\s*\n/, "\n");

  return out;
};

let ok = 0;
let failed = [];
for (const file of targets) {
  const src = fs.readFileSync(file, "utf8");
  const out = stripShell(src);
  if (out === null || out === src) {
    failed.push(file);
    continue;
  }
  fs.writeFileSync(file, out);
  ok++;
}
console.log(`transformed ${ok}/${targets.length}`);
if (failed.length) {
  console.log("FAILED (no shell match):");
  for (const f of failed) console.log("  " + f);
}
