const fs = require("fs");
const path = require("path");

const fixes = [
  {
    file: "src/app/(app)/projects/new/page.tsx",
    from: "</div>\n      </SidebarInset>\n    </SidebarProvider>\n  );\n}",
    to: "</div>\n      </div>\n  );\n}",
  },
  {
    file: "src/app/(app)/projects/[id]/connections/new/page.tsx",
    from: "</CardContent>\n          </Card>\n        </div>\n      </SidebarInset>\n    </SidebarProvider>\n  );",
    to: "</CardContent>\n          </Card>\n        </div>\n      </div>\n  );",
  },
  {
    file: "src/app/(app)/settings/connections/page.tsx",
    from: "</CardContent>\n          </Card>\n        </div>\n  );\n}",
    to: "</CardContent>\n          </Card>\n        </div>\n      </div>\n  );\n}",
  },
  {
    file: "src/app/(app)/projects/[id]/settings/page.tsx",
    from: "</CardContent>\n              </Card>\n            </>\n          )}\n        </div>\n  );\n}",
    to: "</CardContent>\n              </Card>\n            </>\n          )}\n        </div>\n      </div>\n  );\n}",
  },
  {
    file: "src/app/(app)/projects/[id]/settings/members/page.tsx",
    from: "</CardContent>\n          </Card>\n        </div>\n  );\n}",
    to: "</CardContent>\n          </Card>\n        </div>\n      </div>\n  );\n}",
  },
];

for (const fix of fixes) {
  const abs = path.join(__dirname, "..", fix.file);
  let src = fs.readFileSync(abs, "utf8");
  if (!src.includes(fix.from)) {
    console.log("PATTERN NOT FOUND: " + fix.file);
    continue;
  }
  src = src.replace(fix.from, fix.to);
  fs.writeFileSync(abs, src);
  console.log("fixed " + fix.file);
}
