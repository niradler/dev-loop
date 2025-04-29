import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Trash2 } from "lucide-react";

interface DeleteScriptModalProps {
  scriptName: string;
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (removeFile: boolean) => void;
}

export function DeleteScriptModal({
  scriptName,
  isOpen,
  onClose,
  onConfirm,
}: DeleteScriptModalProps) {
  const [removeFile, setRemoveFile] = useState(false);

  const handleConfirm = () => {
    onConfirm(removeFile);
    setRemoveFile(false);
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Trash2 className="h-5 w-5 text-destructive" />
            Delete Script
          </DialogTitle>
          <DialogDescription>
            Are you sure you want to delete the script "{scriptName}"? This action cannot be undone.
          </DialogDescription>
        </DialogHeader>
        <div className="flex items-center space-x-2">
          <Checkbox
            id="removeFile"
            checked={removeFile}
            onCheckedChange={(checked) => setRemoveFile(checked as boolean)}
          />
          <Label htmlFor="removeFile">
            Also delete the script file from disk
          </Label>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={onClose}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={handleConfirm}>
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
} 