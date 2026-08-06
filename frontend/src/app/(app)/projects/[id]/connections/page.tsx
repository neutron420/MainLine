"use client";

import { useParams } from "next/navigation";
import { useState } from "react";
import {
  ArrowLeft,
  Database,
  Loader2,
  Plus,
  RefreshCw,
  Trash2,
  XCircle,
  CheckCircle2,
  Clock,
} from "lucide-react";


import {  Badge  } from "@/components/ui/badge";
import {  Button  } from "@/components/ui/button";
import {  Input  } from "@/components/ui/input";
import {  Label  } from "@/components/ui/label";
import {  Card, CardContent, CardDescription, CardHeader, CardTitle  } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";















import { useConnections, useCreateConnection, useDeleteConnection, useTestConnection } from "@/lib/api/hooks/use-connections";
import type { Connection , TestConnectionResponse } from "@/lib/gen/project/v1/project_messages_pb";






import {  getApiErrorMessage  } from "@/lib/api/errors";

function StatusBadge({ connection }: { connection: Connection }) {
  const status = connection.connectionStatus;
  if (status === "connected") {
    return (
      <Badge variant="outline" className="gap-1 border-green-500/40 bg-green-500/10 text-green-600">
        <CheckCircle2 className="size-3" />
        Connected
      </Badge>
    );
  }
  if (status === "error") {
    return (
      <Badge variant="outline" className="gap-1 border-red-500/40 bg-red-500/10 text-red-600">
        <XCircle className="size-3" />
        Error
      </Badge>
    );
  }
  return (
    <Badge variant="outline" className="gap-1 text-muted-foreground">
      <Clock className="size-3" />
      Not tested
    </Badge>
  );
}

function formatDate(value: string | undefined): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return "—";
  return d.toLocaleDateString(undefined, {
    day: "numeric",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function ConnectionsPage() {
  const params = useParams<{ id: string }>();
  const projectId = params.id;

  const { data: connections = [], isLoading } = useConnections(projectId);
  const createConn = useCreateConnection();
  const deleteConn = useDeleteConnection(projectId);
  const testConn = useTestConnection();

  const [addOpen, setAddOpen] = useState(false);
  const [form, setForm] = useState({
    name: "",
    host: "",
    port: "5432",
    databaseName: "",
    username: "",
    password: "",
    sslMode: "disable",
  });
  const [formError, setFormError] = useState<string | null>(null);
  const [testingId, setTestingId] = useState<string | null>(null);
  const [testResult, setTestResult] = useState<{
    conn: Connection;
    res: TestConnectionResponse | null;
    error: string | null;
  } | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<Connection | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const update = (key: keyof typeof form) => (e: React.ChangeEvent<HTMLInputElement>) => {
    setForm((prev) => ({ ...prev, [key]: e.target.value }));
    setFormError(null);
  };

  const handleAdd = async () => {
    setFormError(null);
    if (!form.name.trim() || !form.host.trim() || !form.databaseName.trim() || !form.username.trim()) {
      setFormError("Name, host, database and username are required.");
      return;
    }
    const port = Number.parseInt(form.port, 10);
    if (Number.isNaN(port) || port < 1 || port > 65535) {
      setFormError("Port must be a valid number (1-65535).");
      return;
    }
    try {
      await createConn.mutateAsync({
        projectId,
        name: form.name.trim(),
        host: form.host.trim(),
        port,
        databaseName: form.databaseName.trim(),
        username: form.username.trim(),
        password: form.password,
        sslMode: form.sslMode,
      });
      setAddOpen(false);
      setForm({ name: "", host: "", port: "5432", databaseName: "", username: "", password: "", sslMode: "disable" });
    } catch (err) {
      setFormError(getApiErrorMessage(err));
    }
  };

  const handleTest = async (connection: Connection) => {
    setTestingId(connection.id);
    setTestResult(null);
    try {
      const res = await testConn.mutateAsync({ projectId, connectionId: connection.id });
      setTestResult({ conn: connection, res, error: null });
    } catch (err) {
      setTestResult({ conn: connection, res: null, error: getApiErrorMessage(err) });
    } finally {
      setTestingId(null);
    }
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deleteConn.mutateAsync(deleteTarget.id);
      setDeleteTarget(null);
    } catch (err) {
      setDeleteError(getApiErrorMessage(err));
    }
  };

  return (
    <div className="flex flex-1 flex-col gap-6 p-6">
      <div className="flex flex-wrap items-center gap-2">
        <Button variant="outline" size="sm" asChild>
              <a href={`/projects/${projectId}`}>
                <ArrowLeft className="size-4" />
                Back
              </a>
            </Button>
          </div>
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-xl font-semibold">Database Connections</h1>
              <p className="text-sm text-muted-foreground">
                Connect your PostgreSQL database to this project for schema management.
              </p>
            </div>
            <Dialog open={addOpen} onOpenChange={setAddOpen}>
              <DialogTrigger asChild>
                <Button>
                  <Plus className="size-4" />
                  Add Connection
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-md">
                <DialogHeader>
                  <DialogTitle>Add Database Connection</DialogTitle>
                  <DialogDescription>
                    Enter your PostgreSQL connection details. Passwords are encrypted at rest.
                  </DialogDescription>
                </DialogHeader>
                <div className="grid gap-4 py-4">
                  <div className="grid gap-2">
                    <Label htmlFor="conn-name">Name</Label>
                    <Input
                      id="conn-name"
                      placeholder="e.g. Production Neon DB"
                      value={form.name}
                      onChange={update("name")}
                    />
                  </div>
                  <div className="grid grid-cols-3 gap-3">
                    <div className="col-span-2 grid gap-2">
                      <Label htmlFor="conn-host">Host</Label>
                      <Input
                        id="conn-host"
                        placeholder="ep-xxx.us-east-1.aws.neon.tech"
                        value={form.host}
                        onChange={update("host")}
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="conn-port">Port</Label>
                      <Input id="conn-port" inputMode="numeric" value={form.port} onChange={update("port")} />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="grid gap-2">
                      <Label htmlFor="conn-db">Database</Label>
                      <Input
                        id="conn-db"
                        placeholder="neondb"
                        value={form.databaseName}
                        onChange={update("databaseName")}
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="conn-user">Username</Label>
                      <Input
                        id="conn-user"
                        placeholder="neondb_owner"
                        value={form.username}
                        onChange={update("username")}
                      />
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-3">
                    <div className="grid gap-2">
                      <Label htmlFor="conn-pass">Password</Label>
                      <Input
                        id="conn-pass"
                        type="password"
                        placeholder="••••••••"
                        value={form.password}
                        onChange={update("password")}
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label htmlFor="conn-ssl">SSL Mode</Label>
                      <Select value={form.sslMode} onValueChange={(v) => setForm((p) => ({ ...p, sslMode: v }))}>
                        <SelectTrigger id="conn-ssl">
                          <SelectValue placeholder="Select SSL mode" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="disable">disable</SelectItem>
                          <SelectItem value="require">require</SelectItem>
                          <SelectItem value="verify-ca">verify-ca</SelectItem>
                          <SelectItem value="verify-full">verify-full</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  {formError ? <p className="text-sm text-red-600">{formError}</p> : null}
                </div>
                <DialogFooter>
                  <Button variant="outline" onClick={() => setAddOpen(false)}>
                    Cancel
                  </Button>
                  <Button onClick={handleAdd} disabled={createConn.isPending}>
                    {createConn.isPending ? <Loader2 className="size-4 animate-spin" /> : null}
                    Add Connection
                  </Button>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </div>

          <Card>
            <CardHeader className="pb-3">
              <CardTitle className="text-base">Connections</CardTitle>
              <CardDescription>
                {isLoading
                  ? "Loading connections…"
                  : connections.length === 0
                    ? "No connections yet — add your first database above."
                    : `${connections.length} connection${connections.length > 1 ? "s" : ""}`}
              </CardDescription>
            </CardHeader>
            <CardContent>
              {isLoading ? (
                <div className="flex items-center justify-center gap-2 py-10 text-sm text-muted-foreground">
                  <Loader2 className="size-4 animate-spin" />
                  Loading…
                </div>
              ) : connections.length === 0 ? (
                <div className="flex flex-col items-center justify-center gap-3 py-12 text-center">
                  <div className="flex size-12 items-center justify-center rounded-full bg-muted">
                    <Database className="size-6 text-muted-foreground" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">No connections yet</p>
                    <p className="text-sm text-muted-foreground">
                      Add a PostgreSQL connection to start managing schemas.
                    </p>
                  </div>
                  <Button variant="outline" onClick={() => setAddOpen(true)}>
                    <Plus className="size-4" />
                    Add Connection
                  </Button>
                </div>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Name</TableHead>
                      <TableHead>Endpoint</TableHead>
                      <TableHead>Database</TableHead>
                      <TableHead>SSL</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Last Connected</TableHead>
                      <TableHead className="text-right">Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {connections.map((connection) => (
                      <TableRow key={connection.id}>
                        <TableCell>
                          <div className="flex items-center gap-2 font-medium">
                            <Database className="size-4 text-muted-foreground" />
                            {connection.name}
                          </div>
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {connection.host}:{connection.port}
                        </TableCell>
                        <TableCell className="text-sm">{connection.databaseName}</TableCell>
                        <TableCell>
                          <code className="rounded bg-muted px-1.5 py-0.5 text-xs">
                            {connection.sslMode || "disable"}
                          </code>
                        </TableCell>
                        <TableCell>
                          <StatusBadge connection={connection} />
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(connection.lastConnectedAt)}
                        </TableCell>
                        <TableCell>
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="sm"
                              onClick={() => handleTest(connection)}
                              disabled={testingId === connection.id}
                            >
                              {testingId === connection.id ? (
                                <Loader2 className="size-4 animate-spin" />
                              ) : (
                                <RefreshCw className="size-4" />
                              )}
                              Test
                            </Button>
                            <Button
                              variant="ghost"
                              size="sm"
                              className="text-red-600 hover:text-red-600"
                              onClick={() => {
                                setDeleteError(null);
                                setDeleteTarget(connection);
                              }}
                            >
                              <Trash2 className="size-4" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              )}
            </CardContent>
          </Card>

      <Dialog open={testResult !== null} onOpenChange={(o) => { if (!o) setTestResult(null); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>
              {testResult?.res?.success ? "Connection Successful" : "Connection Failed"}
            </DialogTitle>
            <DialogDescription>
              {testResult?.conn.name} ({testResult?.conn.host}:{testResult?.conn.port})
            </DialogDescription>
          </DialogHeader>
          {testResult?.res?.success ? (
            <div className="space-y-2 text-sm">
              <div className="flex items-center gap-2 text-green-600">
                <CheckCircle2 className="size-4" />
                Connected successfully
              </div>
              <div className="flex justify-between border-t pt-2 text-muted-foreground">
                <span>Latency</span>
                <span className="font-medium text-foreground">{testResult.res.latencyMs} ms</span>
              </div>
              <div className="flex justify-between text-muted-foreground">
                <span>Server</span>
                <span className="font-medium text-foreground">{testResult.res.serverVersion || "—"}</span>
              </div>
            </div>
          ) : (
            <div className="flex items-start gap-2 text-sm text-red-600">
              <XCircle className="mt-0.5 size-4 shrink-0" />
              {testResult?.res?.error || "Could not connect to the database."}
            </div>
          )}
          <DialogFooter>
            <Button onClick={() => setTestResult(null)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteTarget !== null} onOpenChange={(o) => { if (!o) setDeleteTarget(null); }}>
        <DialogContent className="sm:max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete connection?</DialogTitle>
            <DialogDescription>
              This will permanently remove "{deleteTarget?.name}". This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          {deleteError ? <p className="text-sm text-red-600">{deleteError}</p> : null}
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={deleteConn.isPending}>
              {deleteConn.isPending ? <Loader2 className="size-4 animate-spin" /> : null}
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
