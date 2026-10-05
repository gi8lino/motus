import { Tab, Tabs } from "@mui/material";
import type { View } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

type NavTabsProps = {
  view: View;
  views: View[];
  onSelect: (next: View) => void;
  onPreload: (next: View) => void;
};

const LABELS: Record<View, string> = {
  login: UI_TEXT.nav.login,
  training: UI_TEXT.nav.training,
  workouts: UI_TEXT.nav.workouts,
  exercises: UI_TEXT.nav.exercises,
  history: UI_TEXT.nav.history,
  profile: UI_TEXT.nav.profile,
  admin: UI_TEXT.nav.admin,
};

export function NavTabs({ view, views, onSelect, onPreload }: NavTabsProps) {
  return (
    <Tabs
      value={view}
      onChange={(_event, nextValue: View) => onSelect(nextValue)}
      variant="scrollable"
      allowScrollButtonsMobile
      sx={{
        display: { xs: "none", lg: "flex" },
        minHeight: 0,
        "& .MuiTabs-flexContainer": {
          gap: 0.75,
        },
        "& .MuiTab-root": {
          minHeight: 44,
          px: 1.5,
          borderRadius: 99,
          color: "text.secondary",
        },
        "& .Mui-selected": {
          color: "text.primary",
          bgcolor: "action.selected",
        },
        "& .MuiTabs-indicator": {
          height: 0,
        },
      }}
    >
      {views.map((nextView) => (
        <Tab
          key={nextView}
          value={nextView}
          label={LABELS[nextView]}
          onFocus={() => onPreload(nextView)}
          onMouseEnter={() => onPreload(nextView)}
        />
      ))}
    </Tabs>
  );
}
