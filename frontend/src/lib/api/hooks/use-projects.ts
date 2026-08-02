"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { projectClient } from "@/lib/api/clients";
import type { Project } from "@/lib/gen/project/v1/project_messages_pb";

const DEFAULT_PAGE_SIZE = 50;

export function useProjects() {
  return useQuery({
    queryKey: ["projects"],
    queryFn: async () => {
      const res = await projectClient.listProjects({
        pageSize: DEFAULT_PAGE_SIZE,
      });
      return res.projects;
    },
    staleTime: 30_000,
  });
}

export function useProject(projectId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId],
    queryFn: async () => {
      if (!projectId) throw new Error("Missing project id");
      const res = await projectClient.getProject({ id: projectId });
      return res.project ?? null;
    },
    enabled: Boolean(projectId),
    staleTime: 30_000,
  });
}

export function useCreateProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      name: string;
      description?: string;
      visibility?: string;
      template?: string;
    }) => {
      const res = await projectClient.createProject({
        name: input.name,
        description: input.description ?? "",
        visibility: input.visibility ?? "private",
        template: input.template ?? "blank",
      });
      if (!res.project) throw new Error("Create failed: no project returned");
      return res.project;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useUpdateProject(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: {
      name?: string;
      description?: string;
      visibility?: string;
    }) => {
      const res = await projectClient.updateProject({
        id: projectId,
        name: input.name ?? "",
        description: input.description ?? "",
        visibility: input.visibility ?? "",
      });
      return res.project ?? null;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects", projectId] });
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

export function useDeleteProject() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (projectId: string) => {
      await projectClient.deleteProject({ id: projectId });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects"] });
    },
  });
}

// ── Members ──

export function useMembers(projectId: string | undefined) {
  return useQuery({
    queryKey: ["projects", projectId, "members"],
    queryFn: async () => {
      if (!projectId) throw new Error("Missing project id");
      const res = await projectClient.listMembers({ projectId });
      return res.members;
    },
    enabled: Boolean(projectId),
    staleTime: 30_000,
  });
}

export function useAddMember(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { userId?: string; email?: string; role: string }) => {
      await projectClient.addMember({
        projectId,
        userId: input.userId ?? "",
        email: input.email ?? "",
        role: input.role,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects", projectId, "members"] });
      queryClient.invalidateQueries({ queryKey: ["projects", projectId] });
    },
  });
}

export function useUpdateMemberRole(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { userId: string; role: string }) => {
      await projectClient.updateMemberRole({
        projectId,
        userId: input.userId,
        role: input.role,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects", projectId, "members"] });
      queryClient.invalidateQueries({ queryKey: ["projects", projectId] });
    },
  });
}

export function useRemoveMember(projectId: string) {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (userId: string) => {
      await projectClient.removeMember({ projectId, userId });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects", projectId, "members"] });
      queryClient.invalidateQueries({ queryKey: ["projects", projectId] });
    },
  });
}

export type { Project };
