#!/bin/sh
# Idempotent Theme Unit Test import + write org-seed.json for Gate 3 matrix.
# Env:
#   LAB_UNIT_TEST_XML   (default /lab/fixtures/themeunittestdata.wordpress.xml)
#   LAB_SEED_OUT        (default /var/www/html/wp-content/lab-org-seed.json)
#   LAB_SEED_HOST_OUT   (optional host-mounted path, e.g. /lab/seed-out/org-seed.json)
set -eu

XML="${LAB_UNIT_TEST_XML:-/lab/fixtures/themeunittestdata.wordpress.xml}"
SEED_OUT="${LAB_SEED_OUT:-/var/www/html/wp-content/lab-org-seed.json}"
SEED_HOST_OUT="${LAB_SEED_HOST_OUT:-/lab/seed-out/org-seed.json}"

if ! wp core is-installed --allow-root 2>/dev/null; then
  echo "seed: wp not installed" >&2
  exit 0
fi

if [ ! -f "$XML" ]; then
  echo "seed: XML missing at $XML" >&2
  exit 0
fi

if ! wp option get lab_unit_test_imported --allow-root >/dev/null 2>&1; then
  wp plugin install wordpress-importer --activate --force --allow-root >/dev/null 2>&1 || true
  wp import "$XML" --authors=create --allow-root >/dev/null 2>&1 || true
  wp option update lab_unit_test_imported 1 --allow-root >/dev/null 2>&1 || true
  wp rewrite structure '/%postname%/' --allow-root >/dev/null 2>&1 || true
  wp rewrite flush --hard --allow-root >/dev/null 2>&1 || true
fi

# Gate 4 needs primary menu for desktop nav + mobile sheet.
MENU_ID="$(wp menu list --fields=term_id,name --format=csv --allow-root 2>/dev/null | awk -F, 'NR>1 && tolower($2) ~ /short/ {print $1; exit}')"
if [ -z "${MENU_ID:-}" ]; then
  MENU_ID="$(wp menu list --fields=term_id --format=csv --allow-root 2>/dev/null | awk -F, 'NR==2 {print $1; exit}')"
fi
if [ -n "${MENU_ID:-}" ]; then
  wp menu location assign "$MENU_ID" primary --allow-root >/dev/null 2>&1 || true
fi

ATTACH_ID="$(wp post list --post_type=attachment --post_mime_type=image --fields=ID --format=ids --allow-root 2>/dev/null | awk '{print $1}')"
POST_ID="$(wp post list --post_type=post --post_status=publish --fields=ID --format=ids --allow-root 2>/dev/null | awk '{print $1}')"
PAGE_ID="$(wp post list --post_type=page --post_status=publish --fields=ID --format=ids --allow-root 2>/dev/null | awk '{print $1}')"
TAG_SLUG="$(wp term list post_tag --field=slug --allow-root 2>/dev/null | head -1 | tr -d '\r')"
CAT_ID="$(wp term list category --field=term_id --allow-root 2>/dev/null | head -1 | tr -d '\r')"

[ -n "${POST_ID:-}" ] || POST_ID=1
[ -n "${PAGE_ID:-}" ] || PAGE_ID=2
[ -n "${CAT_ID:-}" ] || CAT_ID=1
[ -n "${TAG_SLUG:-}" ] || TAG_SLUG=test

export ATTACH_ID POST_ID PAGE_ID CAT_ID TAG_SLUG SEED_OUT SEED_HOST_OUT
php -r '
$seed = [
  "attachmentId" => (string) getenv("ATTACH_ID"),
  "postId" => (string) (getenv("POST_ID") ?: "1"),
  "pageId" => (string) (getenv("PAGE_ID") ?: "2"),
  "catId" => (string) (getenv("CAT_ID") ?: "1"),
  "tagSlug" => (string) (getenv("TAG_SLUG") ?: "test"),
  "imported" => true,
];
$json = json_encode($seed, JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES) . "\n";
foreach ([getenv("SEED_OUT"), getenv("SEED_HOST_OUT")] as $path) {
  if (!$path) { continue; }
  $dir = dirname($path);
  if ($dir && !is_dir($dir)) { @mkdir($dir, 0775, true); }
  @file_put_contents($path, $json);
}
echo $json;
'
