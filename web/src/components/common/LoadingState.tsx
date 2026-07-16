import { Skeleton, Stack } from "@mui/material";

export function LoadingState({ rows = 3 }: { rows?: number }) {
  return (
    <Stack spacing={1.25} role="status" aria-label="Loading content">
      {Array.from({ length: rows }, (_, index) => (
        <Skeleton
          key={index}
          variant="rounded"
          height={64}
          animation="wave"
          sx={{ borderRadius: 3 }}
        />
      ))}
    </Stack>
  );
}
