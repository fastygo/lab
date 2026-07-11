/** Stylelint: fatal parse + forbidden legacy CSS only (utility CSS noise ignored). */
export default {
  extends: [],
  rules: {
    // Intentionally empty — we rely on CssSyntaxError + custom entrypoint bans.
  },
  ignoreFiles: ["**/node_modules/**", "**/vendor/**"],
};
