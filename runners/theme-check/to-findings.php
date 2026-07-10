<?php
/**
 * Convert `wp theme-check run --format=json` rows into Lab findings JSON.
 * Input stdin: [ {"type":"REQUIRED","value":"..."}, ... ]
 * Env: LAB_GATE_ID, LAB_CHECK_ID, LAB_TARGET_URL, LAB_THEME_SLUG
 */
declare(strict_types=1);

$raw = stream_get_contents(STDIN);
$data = json_decode($raw ?: '[]', true);
if (!is_array($data)) {
    echo json_encode([
        'findings' => [[
            'code' => 'org.themecheck.parse_failed',
            'gate' => getenv('LAB_GATE_ID') ?: '',
            'check' => getenv('LAB_CHECK_ID') ?: '',
            'severity' => 'high',
            'message' => 'invalid theme-check JSON',
            'target' => getenv('LAB_TARGET_URL') ?: '',
        ]],
    ], JSON_UNESCAPED_SLASHES);
    exit(0);
}

$gate = getenv('LAB_GATE_ID') ?: '';
$check = getenv('LAB_CHECK_ID') ?: '';
$target = getenv('LAB_TARGET_URL') ?: '';
$slug = getenv('LAB_THEME_SLUG') ?: '';

$findings = [];
$required = 0;
$warning = 0;
$info = 0;

foreach ($data as $row) {
    if (!is_array($row)) {
        continue;
    }
    $type = strtoupper(trim((string) ($row['type'] ?? '')));
    $value = trim((string) ($row['value'] ?? ''));
    if ($value === '') {
        continue;
    }

    $severity = 'info';
    $code = 'org.themecheck.info';
    if (str_contains($type, 'REQUIRED') || $type === 'ERROR' || str_contains($type, 'REQUIRED')) {
        $severity = 'high';
        $code = 'org.themecheck.required';
        $required++;
    } elseif (str_contains($type, 'WARNING')) {
        $severity = 'medium';
        $code = 'org.themecheck.warning';
        $warning++;
    } elseif (str_contains($type, 'RECOMMENDED') || str_contains($type, 'INFO')) {
        $severity = 'low';
        $code = 'org.themecheck.recommended';
        $info++;
    } else {
        $info++;
    }

    $findings[] = [
        'code' => $code,
        'gate' => $gate,
        'check' => $check,
        'severity' => $severity,
        'message' => ($type !== '' ? "[{$type}] " : '') . $value,
        'target' => $target,
        'evidence' => [
            'type' => $type,
            'theme' => $slug,
        ],
    ];
}

if ($findings === []) {
    $findings[] = [
        'code' => 'org.themecheck.ok',
        'gate' => $gate,
        'check' => $check,
        'severity' => 'info',
        'message' => 'Theme Check returned no messages for ' . ($slug !== '' ? $slug : 'active theme'),
        'target' => $target,
        'evidence' => ['theme' => $slug],
    ];
} elseif ($required === 0) {
    // Summary finding when only warnings/info
    array_unshift($findings, [
        'code' => 'org.themecheck.no_required',
        'gate' => $gate,
        'check' => $check,
        'severity' => 'info',
        'message' => sprintf('Theme Check: 0 required, %d warning, %d other', $warning, $info),
        'target' => $target,
        'evidence' => [
            'theme' => $slug,
            'required' => '0',
            'warning' => (string) $warning,
        ],
    ]);
}

echo json_encode(['findings' => $findings], JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE);
