import { withBasePath } from "../../utils/basePath";
import { UI_TEXT } from "../../utils/uiText";

// brandLogoSrc resolves the Motus icon URL with the configured base path.
const brandLogoSrc = withBasePath("/brand.svg");

// BrandHeader keeps the brand consistent across early-return layouts.
export function BrandHeader() {
  return (
    <header className="topbar">
      <h1 className="brand">
        <img
          className="brand-logo"
          src={brandLogoSrc}
          alt={UI_TEXT.accessibility.brandAlt}
        />
      </h1>
    </header>
  );
}
