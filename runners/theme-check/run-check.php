<?php
/**
 * Headless Theme Check via plugin APIs (fallback when wp theme-check CLI is absent).
 * Invoked: wp eval-file /runner/run-check.php
 * Env: LAB_THEME_SLUG
 * Stdout: JSON array of {type,value} rows.
 */
declare(strict_types=1);

$slug = getenv('LAB_THEME_SLUG') ?: '';
if ($slug === '') {
    $slug = (string) get_option('stylesheet');
}

$theme = wp_get_theme($slug);
if (!$theme->exists()) {
    echo '[]';
    return;
}

$plugin_dir = WP_PLUGIN_DIR . '/theme-check';
if (!is_dir($plugin_dir)) {
    echo '[]';
    return;
}

// Ensure plugin bootstrap (loads $themechecks).
if (!function_exists('run_themechecks_against_theme')) {
    $main = $plugin_dir . '/theme-check.php';
    if (is_readable($main)) {
        include_once $main;
    }
    if (!function_exists('run_themechecks_against_theme') && is_readable($plugin_dir . '/checkbase.php')) {
        include_once $plugin_dir . '/checkbase.php';
        include_once $plugin_dir . '/main.php';
    }
}

if (!function_exists('run_themechecks_against_theme')) {
    echo '[]';
    return;
}

// Suppress HTML UI output from checks.
ob_start();
run_themechecks_against_theme($theme, $slug);
ob_end_clean();

global $themechecks;
$rows = [];
if (!empty($themechecks) && is_array($themechecks)) {
    foreach ($themechecks as $check) {
        if (!is_object($check) || !method_exists($check, 'getError')) {
            continue;
        }
        $errors = $check->getError();
        if (!is_array($errors)) {
            $errors = $errors ? [(string) $errors] : [];
        }
        foreach ($errors as $err) {
            $html = (string) $err;
            $text = trim(html_entity_decode(wp_strip_all_tags($html), ENT_QUOTES | ENT_HTML5, 'UTF-8'));
            if ($text === '') {
                continue;
            }
            $type = 'INFO';
            if (preg_match('/\bREQUIRED\b/i', $html) || preg_match('/\bREQUIRED\b/i', $text)) {
                $type = 'REQUIRED';
            } elseif (preg_match('/\bWARNING\b/i', $html) || preg_match('/\bWARNING\b/i', $text)) {
                $type = 'WARNING';
            } elseif (preg_match('/\bRECOMMENDED\b/i', $html) || preg_match('/\bRECOMMENDED\b/i', $text)) {
                $type = 'RECOMMENDED';
            } elseif (preg_match('/\bINFO\b/i', $html) || preg_match('/\bINFO\b/i', $text)) {
                $type = 'INFO';
            }
            // Drop leading "REQUIRED:" etc. from value for cleaner messages.
            $value = preg_replace('/^\s*(REQUIRED|WARNING|RECOMMENDED|INFO)\s*:?\s*/i', '', $text) ?: $text;
            $rows[] = ['type' => $type, 'value' => $value];
        }
    }
}

echo json_encode($rows, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
