import { Box, ButtonBase, Stack, Typography } from "@mui/material";
import { withBasePath } from "../../utils/basePath";
import { UI_TEXT } from "../../utils/uiText";

const brandLogoSrc = withBasePath("/brand.svg");

type BrandHeaderProps = {
  auth?: boolean;
  onClick?: () => void;
};

// BrandHeader renders the Motus brand for application and authentication shells.
export function BrandHeader({ auth = false, onClick }: BrandHeaderProps) {
  const logo = (
    <Box
      component="img"
      src={brandLogoSrc}
      alt={UI_TEXT.accessibility.brandAlt}
      sx={{
        width: auth ? 140 : 142,
        height: auth ? 34 : 42,
        objectFit: "contain",
        flexShrink: 0,
        filter: auth ? "none" : "drop-shadow(0 10px 24px rgba(0, 0, 0, 0.18))",
      }}
    />
  );

  return (
    <Stack
      direction="row"
      spacing={auth ? 0 : 1.5}
      sx={{ alignItems: "center", justifyContent: auth ? "center" : undefined }}
    >
      {onClick ? (
        <ButtonBase
          aria-label="Go to training"
          onClick={onClick}
          sx={{ borderRadius: 1, p: 0.5 }}
        >
          {logo}
        </ButtonBase>
      ) : (
        logo
      )}

      {!auth ? (
        <Box>
          <Typography variant="caption" color="text.secondary">
            Training cockpit
          </Typography>
        </Box>
      ) : null}
    </Stack>
  );
}
