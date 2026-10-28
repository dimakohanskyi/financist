import { gql, toId } from "@/api/client";
import type { Tag } from "@/types";

export const mapTag = (t: { id: string; name: string }): Tag => ({ id: toId(t.id), name: t.name });

export async function fetchTags() {
  const data = await gql<{ tags: { id: string; name: string }[] }>(`{ tags { id name } }`);
  return data.tags.map(mapTag);
}

export async function createTag(name: string) {
  const data = await gql<{ createTag: { id: string; name: string } }>(
    `mutation($name: String!) { createTag(name: $name) { id name } }`,
    { name }
  );
  return mapTag(data.createTag);
}

export async function deleteTag(id: number) {
  await gql(`mutation($id: ID!) { deleteTag(id: $id) }`, { id });
}
