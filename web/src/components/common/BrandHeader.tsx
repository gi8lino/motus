import { Box, Stack, Typography } from "@mui/material";
import { withBasePath } from "../../utils/basePath";
import { UI_TEXT } from "../../utils/uiText";

const brandLogoSrc = withBasePath("/brand.svg");

export function BrandHeader() {
  return (
    <Stack direction="row" spacing={1.5} sx={{ alignItems: "center" }}>
      <Box
        component="img"
        src={brandLogoSrc}
        alt={UI_TEXT.accessibility.brandAlt}
        sx={{
          width: 42,
          height: 42,
          objectFit: "cover",
          objectPosition: "left",
          flexShrink: 0,
          filter: "drop-shadow(0 10px 24px rgba(0, 0, 0, 0.18))",
        }}
      />

      <Box>
        <Typography variant="h5" sx={{ lineHeight: 1, fontWeight: 800 }}>
          Motus
        </Typography>
        <Typography variant="caption" color="text.secondary">
          Training cockpit
        </Typography>
      </Box>
    </Stack>
  );
}
