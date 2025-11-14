#!/bin/bash
# Script untuk reset WhatsApp session dan fix database issues

echo "🔧 Resetting WhatsApp Session Database..."

# Stop bot jika sedang running
echo "1️⃣ Stopping bot processes..."
pkill -f "./bot" 2>/dev/null || true
pkill -f "main.go" 2>/dev/null || true
sleep 2

# Backup existing session databases
echo "2️⃣ Backing up existing session databases..."
mkdir -p backup_sessions
if [ -f "data/session.db" ]; then
    cp data/session.db backup_sessions/session_backup_$(date +%Y%m%d_%H%M%S).db
    echo "   ✅ Backed up data/session.db"
fi

if [ -f "simple_session.db" ]; then
    cp simple_session.db backup_sessions/simple_session_backup_$(date +%Y%m%d_%H%M%S).db
    echo "   ✅ Backed up simple_session.db"
fi

if [ -f "visual_session.db" ]; then
    cp visual_session.db backup_sessions/visual_session_backup_$(date +%Y%m%d_%H%M%S).db
    echo "   ✅ Backed up visual_session.db"
fi

# Remove corrupted session databases
echo "3️⃣ Removing corrupted session databases..."
rm -f data/session.db
rm -f simple_session.db
rm -f visual_session.db
rm -f data/qrcode.png
rm -f cmd/qrcode.png
rm -f data/qr_temp.png

echo "   ✅ Removed all session databases"

# Clean temporary files
echo "4️⃣ Cleaning temporary files..."
rm -f data/*.png
rm -f *.png
rm -f /tmp/whatsapp_*

echo "   ✅ Cleaned temporary files"

# Create fresh data directory
echo "5️⃣ Setting up fresh environment..."
mkdir -p data
chmod 755 data

echo "   ✅ Created fresh data directory"

echo ""
echo "🎉 WhatsApp session reset complete!"
echo ""
echo "📋 What was done:"
echo "   • Backed up old session databases"
echo "   • Removed corrupted session files"
echo "   • Cleaned temporary files"
echo "   • Created fresh data directory"
echo ""
echo "🚀 Next steps:"
echo "   1. Start bot: ./bot"
echo "   2. Go to dashboard: http://localhost:1462"
echo "   3. Use WhatsApp Pairing tab for fresh QR scan"
echo ""
echo "💡 Tips:"
echo "   • Use dashboard QR pairing for best experience"
echo "   • Don't logout/login repeatedly (can cause corruption)"
echo "   • If issues persist, restart bot completely"
echo ""