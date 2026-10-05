import AddCircleIcon from "@mui/icons-material/AddCircle";
import {
  Box,
  Button,
  Card,
  CardContent,
  Stack,
  Typography,
} from "@mui/material";

type EmptyStateProps = {
  title: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
};

export function EmptyState({
  title,
  description,
  actionLabel,
  onAction,
}: EmptyStateProps) {
  return (
    <Card variant="outlined" role="status" sx={{ borderStyle: "dashed" }}>
      <CardContent sx={{ py: { xs: 4, md: 5 }, textAlign: "center" }}>
        <Stack
          spacing={1.25}
          sx={{ maxWidth: 420, mx: "auto", alignItems: "center" }}
        >
          <Box
            sx={{
              display: "grid",
              placeItems: "center",
              width: 56,
              height: 56,
              borderRadius: "50%",
              bgcolor: "primary.main",
              color: "primary.contrastText",
            }}
          >
            <AddCircleIcon sx={{ fontSize: 30 }} />
          </Box>
          <Typography variant="h6">{title}</Typography>
          {description ? (
            <Typography variant="body2" color="text.secondary">
              {description}
            </Typography>
          ) : null}
          {actionLabel && onAction ? (
            <Button
              variant="contained"
              size="large"
              onClick={onAction}
              sx={{ mt: 1 }}
            >
              {actionLabel}
            </Button>
          ) : null}
        </Stack>
      </CardContent>
    </Card>
  );
}
