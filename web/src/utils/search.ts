export type SearchableItem = {
  label: string;
  searchText?: string;
};

export function matchesSearch(item: SearchableItem, query: string): boolean {
  const normalized = query.trim().toLowerCase();
  if (!normalized) return true;
  return `${item.label} ${item.searchText ?? ""}`
    .toLowerCase()
    .includes(normalized);
}
