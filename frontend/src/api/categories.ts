import { compact, gql, toId, toIdOrNull } from "@/api/client";
import type { Category, CategoryInput, CategoryKind, CategoryUpdateInput } from "@/types";

export interface GqlCategory {
  id: string;
  name: string;
  kind: CategoryKind;
  icon: string | null;
  color: string;
  sortOrder: number;
  isDefault: boolean;
  parentId: string | null;
}

export const CATEGORY_FIELDS = "id name kind icon color sortOrder isDefault parentId";

export const mapCategory = (c: GqlCategory): Category => ({
  id: toId(c.id),
  name: c.name,
  kind: c.kind,
  icon: c.icon,
  color: c.color,
  sort_order: c.sortOrder,
  is_default: c.isDefault,
  parent_id: toIdOrNull(c.parentId),
});

export async function fetchCategories() {
  const data = await gql<{ categories: GqlCategory[] }>(`{ categories { ${CATEGORY_FIELDS} } }`);
  return data.categories.map(mapCategory);
}

export async function createCategory(input: CategoryInput) {
  const data = await gql<{ createCategory: GqlCategory }>(
    `mutation($input: CreateCategoryInput!) { createCategory(input: $input) { ${CATEGORY_FIELDS} } }`,
    {
      input: compact({
        name: input.name,
        kind: input.kind,
        icon: input.icon || undefined,
        color: input.color,
        sortOrder: input.sort_order,
        parentId: input.parent_id ?? undefined,
      }),
    }
  );
  return mapCategory(data.createCategory);
}

export async function updateCategory(id: number, input: CategoryUpdateInput) {
  const data = await gql<{ updateCategory: GqlCategory }>(
    `mutation($id: ID!, $input: UpdateCategoryInput!) { updateCategory(id: $id, input: $input) { ${CATEGORY_FIELDS} } }`,
    {
      id,
      input: compact({
        name: input.name,
        // "" clears the icon on the backend
        icon: input.icon === undefined ? undefined : (input.icon ?? ""),
        color: input.color,
        sortOrder: input.sort_order,
        parentId: input.parent_id,
      }),
    }
  );
  return mapCategory(data.updateCategory);
}

export async function deleteCategory(id: number) {
  await gql(`mutation($id: ID!) { deleteCategory(id: $id) }`, { id });
}
