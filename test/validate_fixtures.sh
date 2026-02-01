#!/bin/bash
# Test fixture validator - checks log format without running Go tests

FIXTURE_FILE="test/fixtures/sample_logs.txt"
TOTAL=0
VALID=0
SKIPPED=0
ERRORS=0

echo "🧪 Validating log format from $FIXTURE_FILE..."
echo ""

while IFS= read -r line; do
    ((TOTAL++))
    
    # Skip empty lines
    if [[ -z "$line" ]]; then
        ((SKIPPED++))
        continue
    fi
    
    # Skip comments
    if [[ "$line" =~ ^# ]]; then
        ((SKIPPED++))
        continue
    fi
    
    # Split by pipe
    IFS='|' read -ra FIELDS <<< "$line"
    EVENT_TYPE="${FIELDS[0]}"
    
    # Validate based on event type
    case "$EVENT_TYPE" in
        KILL)
            if [[ ${#FIELDS[@]} -lt 11 ]]; then
                echo "❌ Line $TOTAL: KILL event has only ${#FIELDS[@]} fields (need 11)"
                ((ERRORS++))
            else
                ((VALID++))
            fi
            ;;
        DEFLECT)
            if [[ ${#FIELDS[@]} -lt 10 ]]; then
                echo "❌ Line $TOTAL: DEFLECT event has only ${#FIELDS[@]} fields (need 10)"
                ((ERRORS++))
            else
                ((VALID++))
            fi
            ;;
        MATCH_START)
            if [[ ${#FIELDS[@]} -lt 5 ]]; then
                echo "❌ Line $TOTAL: MATCH_START event has only ${#FIELDS[@]} fields (need 5)"
                ((ERRORS++))
            else
                ((VALID++))
            fi
            ;;
        MATCH_END)
            if [[ ${#FIELDS[@]} -lt 6 ]]; then
                echo "❌ Line $TOTAL: MATCH_END event has only ${#FIELDS[@]} fields (need 6)"
                ((ERRORS++))
            else
                ((VALID++))
            fi
            ;;
        *)
            echo "⚠️  Line $TOTAL: Unknown event type '$EVENT_TYPE' (will be skipped by parser)"
            ((SKIPPED++))
            ;;
    esac
done < "$FIXTURE_FILE"

echo ""
echo "📊 Results:"
echo "  Total lines:   $TOTAL"
echo "  Valid events:  $VALID ✅"
echo "  Skipped:       $SKIPPED ⏭️"
echo "  Errors:        $ERRORS ❌"
echo ""

if [[ $ERRORS -gt 0 ]]; then
    echo "❌ Validation FAILED - $ERRORS malformed lines"
    exit 1
else
    echo "✅ Validation PASSED - All event lines have correct field counts!"
    exit 0
fi
