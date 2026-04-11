import {
  alpha,
  createTheme,
  responsiveFontSizes,
  type PaletteMode,
} from "@mui/material/styles";

export function buildAppTheme(mode: PaletteMode) {
  const isDark = mode === "dark";

  const theme = createTheme({
    palette: {
      mode,
      primary: {
        main: isDark ? "#7dd3fc" : "#0f766e",
        light: isDark ? "#bae6fd" : "#14b8a6",
        dark: isDark ? "#0284c7" : "#115e59",
        contrastText: isDark ? "#04141c" : "#f8fffe",
      },
      secondary: {
        main: isDark ? "#f59e0b" : "#c2410c",
        light: isDark ? "#fbbf24" : "#ea580c",
        dark: isDark ? "#d97706" : "#9a3412",
      },
      background: {
        default: isDark ? "#071018" : "#f4efe7",
        paper: isDark ? "#10202d" : "#fffdf8",
      },
      divider: isDark
        ? alpha("#d7ecff", 0.1)
        : alpha("#163047", 0.12),
      text: {
        primary: isDark ? "#eff7ff" : "#102235",
        secondary: isDark ? "#a9bfd2" : "#5c7083",
      },
      success: {
        main: isDark ? "#34d399" : "#15803d",
      },
      warning: {
        main: isDark ? "#fbbf24" : "#b45309",
      },
    },
    shape: {
      borderRadius: 18,
    },
    typography: {
      fontFamily: [
        "Avenir Next",
        "Avenir",
        "Segoe UI",
        "system-ui",
        "sans-serif",
      ].join(", "),
      h1: {
        fontWeight: 800,
        letterSpacing: "-0.04em",
      },
      h2: {
        fontWeight: 800,
        letterSpacing: "-0.03em",
      },
      h3: {
        fontWeight: 700,
        letterSpacing: "-0.02em",
      },
      h4: {
        fontWeight: 700,
        letterSpacing: "-0.02em",
      },
      h5: {
        fontWeight: 700,
      },
      h6: {
        fontWeight: 700,
      },
      button: {
        fontWeight: 700,
        letterSpacing: "0.01em",
        textTransform: "none",
      },
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          body: {
            backgroundImage: isDark
              ? [
                  "radial-gradient(circle at top left, rgba(34, 197, 94, 0.12), transparent 28%)",
                  "radial-gradient(circle at top right, rgba(14, 165, 233, 0.16), transparent 32%)",
                  "linear-gradient(180deg, #08121a 0%, #071018 55%, #06111a 100%)",
                ].join(",")
              : [
                  "radial-gradient(circle at top left, rgba(15, 118, 110, 0.12), transparent 28%)",
                  "radial-gradient(circle at top right, rgba(194, 65, 12, 0.12), transparent 32%)",
                  "linear-gradient(180deg, #f7f3ec 0%, #f4efe7 55%, #efe7dc 100%)",
                ].join(","),
            backgroundAttachment: "fixed",
          },
          "#root": {
            minHeight: "100vh",
          },
        },
      },
      MuiAppBar: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            backdropFilter: "blur(16px)",
          },
        },
      },
      MuiCard: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
            boxShadow: isDark
              ? "0 24px 60px rgba(0, 0, 0, 0.28)"
              : "0 24px 60px rgba(47, 63, 89, 0.12)",
          },
        },
      },
      MuiPaper: {
        styleOverrides: {
          root: {
            backgroundImage: "none",
          },
        },
      },
      MuiButton: {
        defaultProps: {
          disableElevation: true,
        },
        styleOverrides: {
          root: {
            borderRadius: 16,
          },
          sizeLarge: {
            minHeight: 52,
          },
        },
      },
      MuiChip: {
        styleOverrides: {
          root: {
            borderRadius: 999,
            fontWeight: 600,
          },
        },
      },
    },
  });

  return responsiveFontSizes(theme);
}
