import { useState } from "react";
import MenuIcon from "@mui/icons-material/Menu";
import {
  Button,
  Menu,
  MenuItem,
  Tab,
  Tabs,
  useMediaQuery,
  useTheme,
} from "@mui/material";
import type { View } from "../../types";
import { UI_TEXT } from "../../utils/uiText";

type NavTabsProps = {
  view: View;
  views: View[];
  onSelect: (next: View) => void;
};

const LABELS: Record<View, string> = {
  login: UI_TEXT.nav.login,
  train: UI_TEXT.nav.train,
  workouts: UI_TEXT.nav.workouts,
  templates: UI_TEXT.nav.templates,
  exercises: UI_TEXT.nav.exercises,
  history: UI_TEXT.nav.history,
  profile: UI_TEXT.nav.profile,
  admin: UI_TEXT.nav.admin,
};

export function NavTabs({ view, views, onSelect }: NavTabsProps) {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down("md"));
  const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);

  if (isMobile) {
    return (
      <>
        <Button
          color="inherit"
          variant="outlined"
          startIcon={<MenuIcon />}
          aria-label={UI_TEXT.accessibility.navToggle}
          onClick={(event) => setAnchorEl(event.currentTarget)}
          sx={{ justifyContent: "flex-start", borderColor: "divider" }}
        >
          {LABELS[view]}
        </Button>

        <Menu
          anchorEl={anchorEl}
          open={Boolean(anchorEl)}
          onClose={() => setAnchorEl(null)}
        >
          {views.map((nextView) => (
            <MenuItem
              key={nextView}
              selected={view === nextView}
              onClick={() => {
                onSelect(nextView);
                setAnchorEl(null);
              }}
            >
              {LABELS[nextView]}
            </MenuItem>
          ))}
        </Menu>
      </>
    );
  }

  return (
    <Tabs
      value={view}
      onChange={(_event, nextValue: View) => onSelect(nextValue)}
      variant="scrollable"
      allowScrollButtonsMobile
      sx={{
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
        <Tab key={nextView} value={nextView} label={LABELS[nextView]} />
      ))}
    </Tabs>
  );
}
